# gyr - git yaml resolve

similar to how [ko] transforms `ko://<main_go_module>` reference in a piece of
yaml to the reference of a fully built container image in a registry, `gyr`
transforms git references in a piece of yaml to commits SHAs.

for instance

```yaml
foo:
  bar: gyr+gh://cirocosta/gyr#main
```

becomes

```
foo:
  bar: 7f145a7e1057f2f2a099827ab9d7e66a31cd1868
```

## install


```
go get github.com/cirocosta/gyr
```


## usage

`gyr` works by either taking content from stdin

```
echo '---
foo:
  bar: gyr+gh://cirocosta/gyr#main
'| gyr
```

or by opening files specified in positional args (`-` meaning `stdin` - the
default)

```
echo '{foo: {bar: gyr+gh://cirocosta/gyr#main}}' > ./file.yaml
gyr ./file.yaml
```

## supported providers

- github: `gyr+gh://<repository_slug>#<reference>`


## license

see [license](http://www.wtfpl.net/)

[ko]: https://github.com/google/ko
