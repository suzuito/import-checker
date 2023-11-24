package main

import (
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/mod/modfile"
	"gopkg.in/yaml.v3"
)

type Package struct {
	ImportPath string
	Files      []File
}

type Import struct {
	Path string
}

type File struct {
	Path    string
	Imports []Import
}

// ルールの集合
type Rules struct {
	// Import制約ルール
	Rules []Rule `yaml:"rules"`
}

// ルール
// ルールは、あるパッケージがimport可能なImportPathを制約する、という形で記述する。
type Rule struct {
	// ルール名。ルール名は空文字列でも良い。このルール名は、検証にてルール違反が検知された場合、違反したルール名を出力する際に使用される。
	Name string `yaml:"name"`
	// import元のImportPathの指定。言い換えると、importする側のパッケージのImportPath。正規表現で指定すること。
	SrcImportPathPatterns        []string `yaml:"srcImportPathPatterns"`
	SrcImportPathPatternMatchers []regexp.Regexp
	// importが禁止されているImportPathの指定。複数指定可能。正規表現で指定すること。
	ForbiddenImportPathPatterns        []string `yaml:"forbiddenImportPathPatterns"`
	ForbiddenImportPathPatternMatchers []regexp.Regexp
}

type Result struct {
	SrcImportPath string
	Results       []ResultPerFile
}

func (t *Result) HasViolation() bool {
	for _, result := range t.Results {
		if result.HasViolation() {
			return true
		}
	}
	return false
}

type ResultPerFile struct {
	FilePath   string
	Violations []ResultViolation
}

func (t *ResultPerFile) HasViolation() bool {
	return len(t.Violations) > 0
}

type ResultViolation struct {
	RuleName   string
	ImportPath string
}

var headlineUsage = `ルールに従い、1つのGoのモジュール中のあるパッケージが別のパッケージをimportしているかどうか検証する。

ルールはYAML形式で記述される。
`

func usage() {
	// Headline of usage
	fmt.Fprintln(os.Stderr, headlineUsage)
	// Print command line option list
	flag.PrintDefaults()
}

func main() {
	ruleFilePath := ""
	dirPathGoModule := ""
	flag.StringVar(&ruleFilePath, "rule-file", "", "ルールファイルのパス")
	flag.StringVar(&dirPathGoModule, "mod-dir", "./", "Goモジュールのディレクトリパス(デフォルト値はカレントディレクトリ)")
	flag.Parse()
	if ruleFilePath == "" {
		usage()
		os.Exit(1)
	}
	log.SetFlags(0)
	// log.SetOutput(io.Discard)
	run(ruleFilePath, dirPathGoModule)
}

func run(
	ruleFilePath,
	dirPathGoModule string,
) {
	// ルールファイルを読み込む
	ruleFileBytes, err := os.ReadFile(ruleFilePath)
	if err != nil {
		log.Printf("ルールファイルの読み込めませんでした。: %+v\n", err)
		os.Exit(2)
	}
	rules := []*Rule{}
	if err := yaml.Unmarshal(ruleFileBytes, &rules); err != nil {
		log.Printf("ルールファイルをYAML形式として読み込めませんでした。: %+v\n", err)
		os.Exit(2)
	}
	for _, rule := range rules {
		for _, pattern := range rule.SrcImportPathPatterns {
			matcher, err := regexp.Compile(pattern)
			if err != nil {
				log.Printf("不正な正規表現を検知しました。[%s]: %+v\n", pattern, err)
				os.Exit(2)
			}
			rule.SrcImportPathPatternMatchers = append(rule.SrcImportPathPatternMatchers, *matcher)
		}
		for _, pattern := range rule.ForbiddenImportPathPatterns {
			matcher, err := regexp.Compile(pattern)
			if err != nil {
				log.Printf("不正な正規表現を検知しました。[%s]: %+v\n", pattern, err)
				os.Exit(2)
			}
			rule.ForbiddenImportPathPatternMatchers = append(rule.ForbiddenImportPathPatternMatchers, *matcher)
		}
	}

	// go.modを読み込む
	contentGoMod, err := os.ReadFile(path.Join(dirPathGoModule, "go.mod"))
	if err != nil {
		log.Printf("go.modファイルを読み込めませんでした。: %+v\n", err)
		os.Exit(2)
	}
	goModFile, err := modfile.Parse("go.mod", contentGoMod, nil)
	if err != nil {
		log.Printf("go.modファイルをパースできませんでした。: %+v\n", err)
		os.Exit(2)
	}
	modName := goModFile.Module.Mod.Path

	// .goファイルを読み込む
	packages := []Package{}
	if err := filepath.WalkDir(dirPathGoModule, func(dirPathCurrent string, info fs.DirEntry, _ error) error {
		if !info.IsDir() {
			return nil
		}
		fset := token.NewFileSet()
		pkgs, err := parser.ParseDir(fset, dirPathCurrent, func(fi fs.FileInfo) bool { return true }, 0)
		if err != nil {
			return err
		}
		for _, pkg := range pkgs {
			// importする側のパッケージのImportPath
			importPathCurrentPackage := path.Join(modName, dirPathCurrent)
			files := []File{}
			for filePath, goFile := range pkg.Files {
				imports := []Import{}
				for _, imp := range goFile.Imports {
					importPath := strings.Trim(imp.Path.Value, "\"")
					imports = append(imports, Import{Path: importPath})
				}
				files = append(files, File{
					Path:    filePath,
					Imports: imports,
				})
			}
			packages = append(packages, Package{
				ImportPath: importPathCurrentPackage,
				Files:      files,
			})
		}
		return nil
	}); err != nil {
		log.Printf("filepath.Walk関数がエラー終了しました。: %+v\n", err)
		os.Exit(2)
	}

	// Validate
	exitCode := 0
	results := validate(rules, packages)
	for _, result := range results {
		if !result.HasViolation() {
			continue
		}
		log.Printf("## %s\n", result.SrcImportPath)
		log.Println()
		log.Println("下記ファイルに違反があります。")
		log.Println()
		for _, resultPerFile := range result.Results {
			if !result.HasViolation() {
				continue
			}
			exitCode = 3
			log.Printf("- %s\n", resultPerFile.FilePath)
			for _, violation := range resultPerFile.Violations {
				log.Printf("  - \"import %s\"はルール\"%s\"に違反します。\n", violation.ImportPath, violation.RuleName)
			}
		}
		log.Println()
	}
	os.Exit(exitCode)
}

type InvalidRuleError struct {
	SourcePath string
	Message    string
}

func (t *InvalidRuleError) Error() string {
	return fmt.Sprintf("%s : %s", t.SourcePath, t.Message)
}

func validate(rules []*Rule, pkgs []Package) []Result {
	results := []Result{}
	for _, pkg := range pkgs {
		result := Result{
			SrcImportPath: pkg.ImportPath,
			Results:       []ResultPerFile{},
		}
		for _, rule := range rules {
			matched := false
			for _, srcImportPathPatternMatcher := range rule.SrcImportPathPatternMatchers {
				matchedInFor := srcImportPathPatternMatcher.MatchString(pkg.ImportPath)
				if matchedInFor {
					matched = true
				}
			}
			if !matched {
				continue
			}
			for _, pkgFile := range pkg.Files {
				resultPerFile := ResultPerFile{
					FilePath:   pkgFile.Path,
					Violations: []ResultViolation{},
				}
				for _, imp := range pkgFile.Imports {
					for _, forbiddenImportPathPatternMatcher := range rule.ForbiddenImportPathPatternMatchers {
						matched := forbiddenImportPathPatternMatcher.MatchString(imp.Path)
						if matched {
							resultPerFile.Violations = append(resultPerFile.Violations, ResultViolation{
								RuleName:   rule.Name,
								ImportPath: imp.Path,
							})
						}
					}
				}
				result.Results = append(result.Results, resultPerFile)
			}
		}
		results = append(results, result)
	}
	return results
}
