# Hi !

A fun mini project using the divine stack `go;htmx;tailwind;sqlite`.

You need to install go. Project version is 1.24.2

Also requires an access token from [the movies db](https://www.themoviedb.org/)

<br><hr>

In fun mini projects, it's okay to use as few libraries as possible.
All the tailwind is gratefully generated through promptin free, but ever so courteous ChatGPT!
<hr><br>

## How to play

**Tidy your room**
```bash
go mod tidy
```
**Start the circus. Don't forget your token.**
```bash
go run cmd/web-server/main.go -token <your movedb access_token>
```

Navigate to http://localhost:1337/movies on some browser, eg. Firefox.

### *Peace and love* 
