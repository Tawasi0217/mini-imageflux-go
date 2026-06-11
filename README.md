# mini-imageflux-go

mini-imageflux-go は、Goで実装する小規模な画像変換プロキシです。

指定されたオリジン画像を取得し、横幅指定でリサイズしたうえで JPEG として返します。  
変換済み画像はディスクキャッシュへ保存し、同一条件のリクエストでは再変換を行わずにキャッシュから返します。

画像配信基盤に必要な以下の要素を、小規模な実装として検証するためのプロジェクトです。

- 画像取得
- 画像デコード
- リサイズ
- JPEG エンコード
- ディスクキャッシュ
- HTTP キャッシュヘッダ
- ETag
- メトリクス計測
- URL検証
- 巨大画像対策

## Goals

このプロジェクトでは、画像変換プロキシを題材に、以下の要素を実装・検証します。

- オリジン画像を取得する
- リクエストパラメータに応じて画像をリサイズする
- 変換済み画像をディスクキャッシュする
- `Cache-Control` と `ETag` を付与する
- キャッシュヒット率、エラー数、レスポンス時間を計測する
- 不正なURLや巨大画像に対する基本的な制限を入れる
- キャッシュの有無による応答時間の差を観測する

## Motivation

画像配信では、元画像をそのまま返すだけではなく、用途に応じたリサイズ、形式変換、キャッシュ制御、変換コストの管理、メトリクス計測が重要になります。

特に画像変換を伴う配信では、以下のような設計上の論点があります。

- 同じ変換処理を何度も実行しないためのキャッシュ設計
- 変換条件をどのようにキャッシュキーへ反映するか
- ブラウザやCDNに対してどのようなキャッシュヘッダを返すか
- オリジン取得、画像デコード、リサイズ、エンコードのどこがボトルネックになるか
- 巨大画像や不正なURLに対してどう制限をかけるか
- レイテンシやエラー率をどのように観測するか

このプロジェクトでは、これらの要素を段階的に実装し、画像変換プロキシに必要な基本構成を検証します。

## Features

現在の実装状況です。

- [x] Go による HTTP サーバー
- [x] `/image` エンドポイント
- [x] URL 指定によるオリジン画像取得
- [x] JPEG / PNG / GIF 入力のデコード
- [x] 横幅指定によるアスペクト比維持リサイズ
- [x] JPEG 出力
- [ ] WebP 出力
- [x] ディスクキャッシュ
- [x] `Cache-Control` ヘッダ
- [x] `ETag` 対応
- [x] `/metrics` エンドポイント
- [x] キャッシュヒット数・ミス数の計測
- [x] キャッシュヒット率の計測
- [x] エラー数の計測
- [x] レスポンス時間の計測
- [x] キャッシュHIT/MISS別の平均処理時間計測
- [x] HTTPクライアントのタイムアウト
- [x] Content-Length による取得サイズ制限
- [x] 実読み取り量制限
- [x] デコード後のピクセル数上限
- [x] オリジンURL検証
- [x] リダイレクト先URL検証
- [ ] ベンチマーク結果の記録
- [ ] nginx reverse proxy 構成
- [ ] URL 署名

## Architecture

現在の基本構成は以下の通りです。

```txt
Client
  |
  | GET /image?url=...&w=...&format=jpeg
  v
Go Image Proxy
  |
  | 1. Validate request parameters
  | 2. Generate cache key
  | 3. Check disk cache
  | 4. Return cached image if cache hit
  | 5. Fetch origin image if cache miss
  | 6. Decode image
  | 7. Validate decoded image size
  | 8. Resize image by width
  | 9. Encode as JPEG
  | 10. Store converted image to disk cache
  | 11. Return response with cache headers
  v
Disk Cache
  |
  v
Origin Image Server
```

将来的には、前段に nginx を置き、reverse proxy、proxy cache、rate limit なども検証する予定です。

```txt
Client
  |
  v
nginx
  |
  v
Go Image Proxy
  |
  v
Disk Cache / Origin
```

## Request Flow

画像変換リクエストは、以下の流れで処理します。

1. クライアントが変換条件付きでリクエストを送信する
2. リクエストパラメータを検証する
3. オリジンURL、出力幅、出力形式、品質設定からキャッシュキーを生成する
4. 変換済み画像がディスクキャッシュに存在する場合は、キャッシュから返す
5. キャッシュが存在しない場合は、オリジン画像を取得する
6. オリジンURLとリダイレクト先URLを検証する
7. Content-Length と実読み取り量を制限する
8. 画像をデコードする
9. デコード後のピクセル数を検証する
10. 指定された横幅に合わせてアスペクト比を維持したままリサイズする
11. JPEGとしてエンコードする
12. 変換後の画像をディスクキャッシュへ保存する
13. `Cache-Control` と `ETag` を付与してレスポンスを返す
14. リクエスト数、キャッシュヒット率、エラー数、処理時間をメトリクスに記録する

## API

### Image conversion

```txt
GET /image?url={origin_url}&w={width}&format={format}
```

### Example

```txt
GET /image?url=https://example.com/sample.png&w=512&format=jpeg
```

### Parameters

| Parameter | Description | Example | Current support |
|---|---|---:|---|
| `url` | オリジン画像のURL | `https://example.com/sample.png` | required |
| `w` | 出力画像の横幅 | `512` | required, `1` - `4096` |
| `format` | 出力形式 | `jpeg` | currently `jpeg` only |

### Supported formats

| Direction | Format |
|---|---|
| Input decode | JPEG, PNG, GIF |
| Output encode | JPEG |
| Future work | WebP, AVIF |

## Cache Strategy

変換済み画像は、同一条件のリクエストで再変換を行わないようにディスクキャッシュへ保存します。

キャッシュキーは以下の要素から生成します。

- オリジン画像URL
- 出力幅
- 出力形式
- 品質設定

例:

```txt
sha256(url={origin_url}&w={width}&format={format}&q={quality})
```

同じオリジン画像であっても、変換条件が異なる場合は別のキャッシュとして扱います。

### Cache Hit

キャッシュヒット時は、オリジン画像の取得、画像デコード、リサイズ、エンコードを行わず、保存済みの変換結果を返します。

レスポンスには以下のようなヘッダを付与します。

```txt
X-Image-Proxy-Cache: HIT
```

### Cache Miss

キャッシュミス時は、オリジン画像を取得し、指定された条件で変換を行った後、変換済み画像をキャッシュへ保存します。

```txt
X-Image-Proxy-Cache: MISS
```

## HTTP Cache Headers

変換結果には、ブラウザやCDNでのキャッシュを想定してHTTPキャッシュヘッダを付与します。

現在付与しているヘッダ例:

```txt
Cache-Control: public, max-age=31536000, immutable
ETag: "{sha256_of_converted_image}"
```

`ETag` は変換後画像のバイト列から生成します。  
クライアントが `If-None-Match` を送信し、ETagが一致した場合は `304 Not Modified` を返します。

現時点では、ETag判定のために変換処理が走るケースがあります。  
ただしディスクキャッシュにヒットした場合は、キャッシュ済みファイルからETagを生成して判定できます。

## Safety Checks

画像プロキシでは `url` パラメータで任意のURLを指定できるため、最低限の安全性チェックを実装しています。

現在の実装内容は以下です。

- `http` / `https` 以外のスキームを拒否
- `localhost` を拒否
- loopback IP を拒否
- unspecified IP を拒否
- リダイレクト先URLも検証
- リダイレクト回数を制限
- HTTPクライアントにタイムアウトを設定
- `Content-Length` が上限を超える画像を拒否
- `http.MaxBytesReader` による実読み取り量制限
- デコード後のピクセル数上限
- 出力幅 `w` の上限

この実装は基本的な防御であり、完全なSSRF対策ではありません。  
DNS解決後のプライベートIP判定や、より厳密なリダイレクト制御は今後の課題です。

## Metrics

`/metrics` エンドポイントで、Prometheus text formatに近い形式のメトリクスを返します。

現在確認できる値は以下です。

- 総リクエスト数
- キャッシュヒット数
- キャッシュミス数
- キャッシュヒット率
- エラー数
- レスポンス時間の合計
- レスポンス時間の平均
- キャッシュHIT時の平均処理時間
- キャッシュMISS時の平均処理時間

出力例:

```txt
image_proxy_requests_total 14
image_proxy_cache_hits_total 13
image_proxy_cache_misses_total 1
image_proxy_cache_hit_rate 0.9286
image_proxy_errors_total 0
image_proxy_response_time_seconds_total 0.316298
image_proxy_response_time_seconds_avg 0.022593
image_proxy_cache_hit_response_time_seconds_avg 0.000319
image_proxy_cache_miss_response_time_seconds_avg 0.312152
```

この例では、キャッシュHIT時の平均処理時間は約0.319ms、キャッシュMISS時の平均処理時間は約312msです。  
キャッシュによって、オリジン取得・画像デコード・リサイズ・エンコードを省略できていることが分かります。

## Benchmark

現在は `/metrics` による簡易計測まで実装しています。  
今後、`hey` などのHTTP負荷試験ツールを使って、キャッシュの有無によるレイテンシ差を測定します。

| Condition | Avg | Notes |
|---|---:|---|
| Cache miss | 0.312152s | origin fetch + decode + resize + encode + cache write |
| Cache hit | 0.000319s | disk cache |
| WebP conversion | TBD | future work |

測定例:

```bash
hey -n 100 -c 10 "http://localhost:8080/image?url=https://example.com/sample.png&w=512&format=jpeg"
```

## Getting Started

### Requirements

- Go 1.22+
- Git

### Run

```bash
git clone https://github.com/yourname/mini-imageflux-go.git
cd mini-imageflux-go
go run ./cmd/server
```

### Example Request

```bash
curl "http://localhost:8080/image?url=https://example.com/sample.png&w=512&format=jpeg" -o output.jpg
```

### Metrics

```bash
curl "http://localhost:8080/metrics"
```

## Tech Stack

- Go
- `net/http`
- `image`
- `image/jpeg`
- `image/png`
- `image/gif`
- `golang.org/x/image/draw`
- Disk cache
- Prometheus-like text metrics

## Directory Structure

```txt
mini-imageflux-go/
  cmd/
    server/
      main.go
  internal/
    cache/
      disk.go
    fetcher/
      fetcher.go
      url_validate.go
    imageproc/
      decode.go
      encode.go
      resize.go
      size_validate.go
    metrics/
      metrics.go
    server/
      handler.go
      metrics_handler.go
  README.md
  go.mod
  go.sum
```

## Current Status

このプロジェクトは、画像変換プロキシの基本MVPを実装済みです。

現在できることは以下です。

- URLから画像を取得する
- JPEG / PNG / GIF をデコードする
- 横幅指定でリサイズする
- JPEGで返す
- 変換済み画像をディスクキャッシュする
- `Cache-Control` と `ETag` を返す
- `/metrics` で基本的なメトリクスを返す
- キャッシュHIT/MISS別の応答時間を確認する
- URL検証と画像サイズ制限を行う

次の段階では、ベンチマーク結果の記録、READMEの説明強化、WebP対応の検討を行います。

## Known Issues

現時点で把握している課題は以下です。

- WebP / AVIF 変換は未対応
- キャッシュ削除ポリシーが未実装
- 同一画像への同時リクエスト時に重複変換が発生する可能性がある
- DNS解決後のプライベートIP判定は未実装
- キャッシュ破損時の扱いが簡易的
- オリジン取得失敗時のリトライ設計が未実装
- PNG / GIF の透過情報をJPEG出力時に十分考慮していない
- アニメーションGIFは静止画として扱われる可能性がある
- HEADリクエストには未対応
- 本格的なベンチマークは未実施

## Future Work

今後は以下の機能を追加・検証する予定です。

- WebP出力
- AVIF対応
- nginx reverse proxy構成
- nginx proxy cacheの検証
- URL署名
- 期限付きURL
- Range Request対応
- worker poolによる変換処理制御
- rate limit
- pprofによるCPU / memory profiling
- Prometheus / Grafanaによる可視化
- DNS解決後のSSRF対策強化
- キャッシュ削除ポリシー
- TypeScript / Vueによる簡易管理画面

## Notes

このプロジェクトは、画像変換プロキシを題材に、画像配信基盤の基本要素を検証するための実験的な実装です。  
本番運用を目的としたものではありません。
