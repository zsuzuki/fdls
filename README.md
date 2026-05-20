# fdls

`find` の動作を絞った、ファイル一覧取得用のGo CLIです。

## Build

```sh
go build -o fdls .
```

## Usage

```sh
fdls [options] <directory>
```

Options:

- `-path rel|abs`: パス表示形式。既定値は `rel`
- `-match string`: ルートからの相対パスに指定文字列を含むファイルだけを表示
- `-exclude-dir name-or-path`: 指定したディレクトリを探索対象から除外。複数回指定可能
- `-sha256`: SHA256ハッシュ値を表示
- `-date`: 更新日時を表示
- `-size`: サイズをバイト単位で表示
- `-depth N`: 探索階層。`-1` は無限、`0` は指定ディレクトリ直下のみ

出力はタブ区切りです。

```sh
fdls -path abs -size -date -sha256 -depth 2 .
```

```sh
fdls -match ".go" .
```

```sh
fdls -exclude-dir vendor -exclude-dir .git .
```

`-exclude-dir vendor` は任意階層の `vendor` ディレクトリを除外します。`-exclude-dir build/cache` のようにパスを指定した場合は、ルートからの相対パスとして一致したディレクトリだけを除外します。

スペースを含むパスやファイル名もそのまま扱います。Windowsでは引用符で囲むのが確実です。

```powershell
fdls -size "C:\Program Files"
```

未引用で `fdls C:\Program Files` のように渡された場合も、残りの位置引数をスペースで結合して1つのディレクトリとして扱います。
