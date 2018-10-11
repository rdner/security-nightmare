# security-nightmare
A demo for a talk on web-security. **Don't deploy this app. EVER!**

This web-app consists of two HTTP servers listening to `:666` and `:777`.

* `:666` – the most insecure blog post feed ever created by the humanity. Does not escape anything, so there are plenty ways to exploit it.
* `:777` – a demo server needed for CSRF and `window.opener` attacks.

By default, the feed database is pre-filled with examples which can give you some guidance.

## Run

You have to install [go](https://golang.org/doc/install) and [dep](https://github.com/golang/dep) in order to build and run it.

Then to start shooting your legs is simple as:

```bash
mkdir -p $GOPATH/src/github.com
cd $GOPATH/src/github.com
git clone git@github.com:pragmader/security-nightmare.git
cd security-nightmare
make run
```

## MIT License

Copyright 2018 Denis Rechkunov

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

**THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.**
