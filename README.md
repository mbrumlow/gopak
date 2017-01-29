# gopak
Use this program to bundle a files with a go executable. 

## Overview 

To bundle files with a go executable replace all instances of os.Open with gopak.Open and adjust for receiving a io.ReadCloser. 
Then run gopak -file <path to go program> -root <path to root of name space> 
This will append a zip file along with footer to the end of the go binary. 

Files paths are grouped in name spaces. You can have more than one name-space. A name space is a root directory. 

This system was setup so your gopak.Open files will work in a development environment by looking for the files on the local disk if the gopak zip and footer have not been added yet. 

## Example 

Given the following directory structure.

````
project/
├── main.go
└── webroot
    ├── images
    │   └── title.jpg
    └── index.html 
````               

When run from the projects directory gopak.Open("webroot", "images/title.jpg") and gopak.Open("webroot", "index.html") will properly work with and without actually packing the file to the end of the binary. This means that 'go run main.go' will still work. 

## Caveats 

All name spaces should be relative to the directory the program is ran in -- this is required if you wish to run your program without first appending the zip file. For example if you want to run "go run main.go" 


Also not tested in windows. 
