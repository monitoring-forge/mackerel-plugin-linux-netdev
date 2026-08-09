# mackerel-plugin-linux-netdev

Linux の `/proc/net/dev` を収集する Mackerel 用メトリックプラグインです。各ネットワークインターフェースの **エラー数 / ドロップ数 / パケット数** を 1 秒あたりの値で出力します。

## インストール

以下のいずれかの方法でインストールしてください。

- リリースページからバイナリをダウンロードする
- `mkr plugin install monitoring-forge/mackerel-plugin-linux-netdev` を実行する

## 使い方

```
Usage:
  mackerel-plugin-linux-netdev [OPTIONS]

Application Options:
  -v, --version            Show version
      --ignore-interfaces= Regexp for interfaces name to ignore

Help Options:
  -h, --help               Show this help message
```

基本的な実行例です。特別なオプションがなければ、すべてのインターフェース（`lo` を除く）を収集します。

```
$ ./mackerel-plugin-linux-netdev
linux-netdev.errors.all.tx      0       1634886461
linux-netdev.errors.eth0.tx     0       1634886461
linux-netdev.errors.eth0.rx     0       1634886461
linux-netdev.errors.all.rx      0       1634886461
linux-netdev.dropped.eth0.tx    0       1634886461
linux-netdev.dropped.all.tx     0       1634886461
linux-netdev.dropped.eth0.rx    0       1634886461
linux-netdev.dropped.all.rx     0       1634886461
linux-netdev.pps.eth0.tx        1.666667        1634886461
linux-netdev.pps.eth0.rx        1529.333333     1634886461
```

### オプション

#### `--ignore-interfaces`

収集対象から除外するインターフェース名を正規表現で指定します。たとえば `docker` 系や `veth` 系を除外したい場合は次のようにします。

```
$ ./mackerel-plugin-linux-netdev --ignore-interfaces='^(docker|veth|br-)'
```

## 出力メトリック

出力されるメトリックは次の 3 種類です。それぞれ 1 秒あたりの値で出力されます。

| メトリック名 | 意味 |
| --- | --- |
| `linux-netdev.errors.{iface}.tx` / `.rx` | 送信 / 受信時に発生したエラー数 |
| `linux-netdev.dropped.{iface}.tx` / `.rx` | 送信 / 受信時にドロップされたパケット数 |
| `linux-netdev.pps.{iface}.tx` / `.rx` | 送信 / 受信されたパケット数（packets per second） |

`{iface}` には `eth0` などのインターフェース名が入ります。また、`{iface}` が `all` の項目は、収集対象となったすべてのインターフェースの合計値です。

## 注意点

- 初回実行時は前回値が存在しないため、メトリックは出力されません。2 回目以降の実行で差分が計算されます。
- 前回値は一時ファイルとして保存されます。デフォルトでは Mackerel プラグインの作業ディレクトリに保存されます。
- 前回値から 600 秒以上経過している場合、値が古すぎるとしてエラーになります。
- `lo`（ループバック）は常に収集対象外です。

## 開発

テストは次のコマンドで実行できます。

```
$ go test ./...
```