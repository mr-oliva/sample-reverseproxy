# GoでReverseProxyを書く際にデバッグを頑張る話

## 概要
本記事は [Go Advent Calendar 2019](https://qiita.com/advent-calendar/2019/go)の14日目の記事となります。
この1年はけっこうGoを書いたのでネタには困らないはずだったのですが、いざ書くとなるとネタがないものですねｗ

1年で1番ハマったことを書くつもりがあまり思い出せないので直近でハマった**リバプロ**周りのことについて書いていこうと思います。

## サンプルコード
本記事用に以下のrepositoryを用意したのでこちらを使って説明していきます。
[サンプルコード](https://github.com/bookun/sample-reverseproxy)

## Goでリバースプロキシを書く

![simple](https://raw.githubusercontent.com/bookun/sample-reverseproxy/master/images/qiita1.png)

|上記画像上のIP | サンプル上のhost |
|:----: | :---:|
| B | localhost:8081 |

こういった場合は簡単にリバースプロキシを書くことができます

[simple/main.go](https://raw.githubusercontent.com/bookun/sample-reverseproxy/master/simple/main.go)

```go
http.Handle("/", httputil.NewSingleHostReverseProxy("http://localhost:8081"))
```

これだけで大丈夫かと思います。

## ハマった問題

下記の状態のときのリバースプロキシを書く際にドハマリしました。    
(そして挙動を理解しきれておらずヤラカしましたが、今回はその内容には触れません)

![problem](https://raw.githubusercontent.com/bookun/sample-reverseproxy/master/images/qiita2.png)

リバプロ導入したからと言って、従来のリライトの挙動が動作しなくなるのは困ります。   
どういうことをしたいかというと **通信はIP Bに対して行うが、IP Bを持つサーバのApacheが通信を受け取った際に Host Header が Service B1もしくは Service B2のドメインを検知できるようにする** ということを達成したいことになります。

|上記画像上のIP | サンプル上のhost |
|:----: | :---:|
| B | localhost:8081 |

|上記画像上のhostヘッダ | サンプル上のhostヘッダ |
|:----: | :---:|
| Service B1| 特になにもなし |
| Service B2| go.advent.2019.co.jp (架空のものです) |

### まず結果から

詳細は下記のコードを見てください
[problem/main.go](https://raw.githubusercontent.com/bookun/sample-reverseproxy/problem/main.go)

ポイントとしてはDirectorを以下のように設定すると要件を満たすことができます。

```go
director := func(req *http.Request) {
	req.URL.Scheme = scheme
	req.URL.Host = "http://localhost:8081"
	req.Host = "go.advent.2019.co.jp"
}
```

##### サンプルプログラムの実行方法

```
$ docker-compose up -d
$ cd problem
$ go run main.go
```

このあとブラウザで http://localhost:8080/sample_image.png を開くと, http://localhost:8081/sample_image2.png と同じ画像になります。

とりあえず結果だけわかればOKの方はここまでで大丈夫なはずです。

### なんでうまくいくのか?
ま、まぁ、あれかな。    
通信はURLに対して行うけど、Hostヘッダの中身は `req.Host` になるってことだなたぶん。
(次節で確認します）

[mercariさんのテックブログ](https://tech.mercari.com/entry/2018/12/05/105737)にも

> もしproxy先へのリクエストのHostヘッダーを書き換えたい場合はdirectorの中でreq.Hostに代入して書き換えます.

ってあるし。


### Debug してみる

httpリクエストをTraceする必要があると考え、調べてみたところ official blogに[いい記事](https://blog.golang.org/http-tracing)がありました。

下記を追記します

``` Go
//refer to https://blog.golang.org/http-tracing
type transport struct {
	current *http.Request
}

func (t *transport) GotConn(info httptrace.GotConnInfo) {
	fmt.Printf("Connected to %v\n", t.current.URL)
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.current = req
	b, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(b))
	return http.DefaultTransport.RoundTrip(req)
}
```

これらは下記のように呼び出します。

```	Go
t := &transport{}
trace := &httptrace.ClientTrace{
	GotConn: t.GotConn,
}

r = r.WithContext(httptrace.WithClientTrace(r.Context(), trace))

reverse := &httputil.ReverseProxy{Director: director, Transport: t}
reverse.ServeHTTP(w, r)
```


##### サンプルプログラムの実行方法

```
$ docker-compose up -d
$ cd problem
$ DEBUG=1 go run main.go
```

ブラウザにて http://localhost:8080/sample_image.png を開いたときにコンソール上にログが出力されるはずです。

```
Connected to http://localhost:8081/sample_image.png

GET /sample_image.png HTTP/1.1
Host: go.advent.2019.co.jp
User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:71.0) Gecko/20100101 Firefox/71.0
Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8
Accept-Encoding: gzip, deflate
略
```

接続は `http://localhost:8081` に対して行われ、リクエストヘッダ内のHostは `go.advent.2019.co.jp` になってることがわかりますね！

分量的にだいぶ増えてしまうので省略しますが、`httputil.ReverseProxy` からコードを追っていくと、最終的にコネクションは req.URLの情報を使用しているようです。（間違ってたらすみません）

https://github.com/golang/go/blob/master/src/net/http/transport.go#L743

たぶんここ。

もう少し時間あるときに `net/http` のコードをもう少し読んでみようと思います。
