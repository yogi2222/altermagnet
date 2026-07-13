# Batch Altermagnetism Scanner

A Go command-line utility for running [`amcheck`](https://pypi.org/project/amcheck/) across a folder of VASP structure files and identifying files that return `Altermagnet? True` for at least one tested spin configuration.

The program automates interactive `amcheck` prompts, generates spin combinations for selected transition-metal elements, logs all output, and copies matching files into a `trueFiles` folder.

## Features

- Checks whether `amcheck` is installed before running.
- Installs `amcheck` automatically with `pip3 install amcheck` if it is missing.
- Processes every file in a user-provided folder.
- Responds automatically to the primitive cell prompt with `Y`.
- Detects atom counts from `amcheck` output.
- Tests generated `u`/`d` spin configurations for selected elements.
- Uses `nn` for non-target elements.
- Counts how many tested configurations return `Altermagnet? True`.
- Copies files with at least one positive result into a `trueFiles` output directory.
- Writes a full run log to `amcheck_log.txt`.

## Target Elements

The program generates spin combinations for the following elements:
```text
Ti, V, Cr, Mn, Fe, Co, Ni, Cu, Mo, Ru
```
(You can change it in the code anytime)

All other elements are assigned `nn`.

## How It Works

1. The program checks whether `amcheck` is available.
2. If `amcheck` is not installed, it attempts to install it using `pip3`.
3. The user enters a folder path containing VASP structure files.
4. For each file, the program first runs `amcheck` using default/generated `nn` responses to identify relevant elements and atom counts.
5. For target elements, it generates valid spin sequences containing equal numbers of `u` and `d` values.
6. Duplicate configurations related by global spin flip are removed.
7. Each generated configuration is tested with `amcheck`.
8. If any configuration returns `Altermagnet? True`, the source file is copied into `trueFiles`.

## Requirements

- Go
- Python 3
- `pip3`
- `amcheck`
- VASP structure files compatible with `amcheck`

> The program can install `amcheck` automatically, but Python and `pip3` must already be available on your system.

## Installation

Clone this repository:

```bash
git clone <https://github.com/yogi2222/altermagnet>
cd <altermagnet>
```


```bash
ls
```

You should see the main Go file:

```text
main2.go
```

## Usage

Run the program:

```bash
go run main2.go
```

When prompted, enter the path to the folder containing the files you want to check:

```text
Enter folder path: /path/to/vasp/files
```

The program will process each file in that folder.

## Output

If your input folder is:

```text
/path/to/project/inputFiles
```

then the program creates output files one level above that folder:

```text
/path/to/project/trueFiles/
/path/to/project/amcheck_log.txt
```

### `trueFiles/`

Contains copies of all input files where at least one tested spin configuration returned:

```text
Altermagnet? True
```

### `amcheck_log.txt`

Contains detailed output from every `amcheck` run, including:

- processed files
- detected elements
- atom counts
- spin responses sent to `amcheck`
- final altermagnetism results
- total tested combinations
- number of positive configurations

## Example Console Output

```text
amcheck is already installed.
Enter folder path: /path/to/vasp/files
Files in folder:
Sending response: nn
Altermagnet Result: False
Out of 6 combinations, 1 returned Altermagnet? True
Altermagnet found in one or more configurations.
File /path/to/vasp/files/example.vasp copied successfully to: /path/to/trueFiles/example.vasp
```

## Notes

- Folder paths with spaces may not work because the program reads the path using `fmt.Scanln`.
- For target elements with atom counts greater than 24, the file is skipped with a warning.
- Runtime will grow exponentially as magnetic atom counts increase because many spin combinations will be tested.
- Existing `amcheck_log.txt` in the output directory is deleted at the start of each run.
- The program assumes that the `amcheck` command-line prompts match the expected text patterns.

## Project Structure

```text
.
├── main2.go
└── README.md
```

Generated after running:

```text
..
├── amcheck_log.txt
└── trueFiles/
```

## Building a Binary

You can build the program into an executable:

```bash
go build -o amcheck-batch main2.go
```

Then run it:

```bash
./amcheck-batch
```
