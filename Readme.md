# Fastats

Tools for bioinformatics in Go

## Intro

This package contains a large number of utility scripts for more easily writing
bioinformatics programs in Go. Most notably, it contains parsers and printers
for most of the major bioinformatics file types, including fasta, fastq, gff,
gtf, bed, newick, and sam. It also contains a set of executables made with
these tools to do common bioinformatics tasks. For example, the "fastats" executable
prints contiguity statistics (N50, etc.) for fasta files.

## Dependencies

This package depends on several other Go packages maintained by me. See the
file "go.mod" for more info.

## Library installation

To install the latest version of this library, just use:

```sh
go get github.com/jgbaldwinbrown/fastats/pkg
```

Then, to install any of the command line tools, do the following:

```sh
cd fastats
go install cmd/[tool]
```

If that doesn't work, manually install the tool as follows:

```sh
cd fastats
go build cmd/[tool]/[tool].go
cp [tool] /usr/bin/
```
