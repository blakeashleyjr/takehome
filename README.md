# Take Home Project

Challenge: A directory contains multiple files and directories of non-uniform file and directory names. Create a program that traverses a base directory and creates an index file that can be used to quickly lookup files by name, size, and content type.

# Usage

`index-search` is a simple program to index and (optionally) search a directory of files recursively. It creates a .csv file that contains the `Name,Size,Type,Path` of the files in the given directory. Type is a best guess based on the encoding of the data. It then allows the user to search for a keyword, size, type, etc. and will list all the matching entries for the search term. The current implementation only searches metadata not content.

The program offers the following flags

```
-i, --index, Create the index file. 
-d, --directory, The directory to index, required if the --index flag is set. 
-s, --search, The search query to run against the index. An index file must be present in order to search. 
-v, --verbose, Increase the verbosity of logs and error messages, useful for troubleshooting.
```

You can explore the source code yourself in main.go. Test any changes with `go run main.go` and build them when you are ready `go build -o index-search main.go`.

## **Examples**

*Index the test-data directory:* <br>
`./index-search -i -d test_data/`

Which will output something like the following: 
```
"{"Successfully created index file","filename":"index.csv","fileCount":5}"
```

*Search the index from the previous example for all json files:* <br>
`./index-search -s json` <br>

Which will output something like the following: <br>
```
[user1.json 16 application/octet-stream test_data/data/user1.json]
[user2.json 17 application/octet-stream test_data/data/user2.json]
```

*Combine the two previous examples into a single command:* <br>
`./index-search -i -d test_data/ -s json`

Which will output something like the following: <br>
```
{"level":"info","ts":1688587329.809959,"caller":"takehome/main.go:189","msg":"Successfully created index file","filename":"index.csv","fileCount":5}
[user1.json 16 application/octet-stream test_data/data/user1.json]
[user2.json 17 application/octet-stream test_data/data/user2.json]
```
