# sample of reverse proxy with Go

## Get started


![simple](https://raw.githubusercontent.com/bookun/sample-reverseproxy/master/images/qiita1.png)

```
$ docker-compose up -d
$ cd simple
$ go run main.go
```

And then, please access http://localhost:8080/sample_image.png with your browser.

![problem](https://raw.githubusercontent.com/bookun/sample-reverseproxy/master/images/qiita2.png)

```
$ docker-compose up -d
$ cd problem 
$ go run main.go
```

And then, please access http://localhost:8080/sample_image.png with your browser.
