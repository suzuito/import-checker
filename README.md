# import-checker

パッケージの依存関係を静的解析するコマンドを作りました。このコマンドの利用者は、あるパッケージが別のパッケージをimportすることを禁止することを意味するルールを定義します。コマンドは、モジュールがルールを満たしているかどうかを検証します。

ルール定義はYAML形式で記述します。

```yaml
- name: domain層にあるソースコードが、usecase,infra層のパッケージをimportすることを禁止する。
  srcImportPathPatterns:
    # import元のImportPathを正規表現で記載する
    - ^github\.com/suzuito/demo001/domain$
    - ^github\.com/suzuito/demo001/domain/.+$
  forbiddenImportPathPatterns:
    # import元が禁止されているImportPathを正規表現で記載する
    - ^github\.com/suzuito/demo001/usecase$
    - ^github\.com/suzuito/demo001/usecase/.+$
    - ^github\.com/suzuito/demo001/infra$
    - ^github\.com/suzuito/demo001/infra/.+$
```

コマンドを実行します。

```bash
import-checker -rule-file rules.yaml -mod-dir ./
```

モジュール配下にあるGo言語のソースコードがルールを違反していた場合、次のように標準出力へ結果が出力され、コマンドが0以外のステータスで終了します。

```bash
## github.com/suzuito/demo001/domain/foo

下記ファイルに違反があります。

- ./domain/foo/bar.go
  - "import github.com/suzuito/demo001/usecase/hoge"はルール"domain層にあるソースコードが、usecase,infra層のパッケージをimportすることを禁止する。"に違反します。

```

## インストール方法

```bash
go install github.com/suzuito/import-checker
```

## ルールファイルのスキーマ

jsonスキーマで定義しました。

[schema.json](schema.json)
