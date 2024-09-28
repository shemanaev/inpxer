# INPX parser

Library for reading `.inpx` collection files.

Example:

```go
collection, err := inpx.Open("testdata/flibusta_fb2_local.inpx")
if err != nil {
    panic(err)
}
defer collection.Close()

for book := range collection.Stream() {
    fmt.Printf("%v\n", book)
}
```
