# magicx

### rename
```
$ ./bin/magicx rename --help

Usage:
  magicx [OPTIONS] rename [rename-OPTIONS]

The rename command-line fix the numbering file.

Help Options:
  -h, --help      Show this help message

[rename command options]
      -p, --path= Full path (default: /Users/JP17278/Downloads/00022_sansyoku)
      -n, --num=  Suffix number (default: 3)
```

```
./bin/magicx rename --path=xxx
```

### resize
```
$ ./bin/magicx resize --help

Usage:
  magicx [OPTIONS] resize [resize-OPTIONS]

The resize command-line

Help Options:
  -h, --help         Show this help message

[resize command options]
      -p, --path=    Full path (default: /Users/JP17278/Downloads/00022_sansyoku)
      -w, --width=   Limit width (default: 2266)
      -s, --size=    Limit size (kb) (default: 30720)
          --percent= Resize percentages (default: 95.0)
```