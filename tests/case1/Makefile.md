# Testcase 1

## `sayhello:`

```sh
@echo "Hello, $(WHO)!"
```

## `saygoodbye:`

> `: sayhello`

```sh
@echo "Bye for now," `date +%Y/%m/%d`
```

```sh
@echo "See you later."
```

## `variables:`

```makefile
WHO = makedown
```
