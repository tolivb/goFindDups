# goFindDups
Simple tool to scan dirs for duplicate files by using MD5 sums.It will generate the result on STDOUT.

1. The tool WILL NOT MOVE or DELETE anything, it will just output the duplicates and bash commands for convinience.
2. Use absolute paths only!
3. Files with size 0 are considered equal!

## How to use
``` 
./finddups /dir1 /dir2 /dir3
```

## Example

```
./finddups /home/ninja/dir1/ /home/ninja/dir2/
Starting with 3 workers

ERROR: open /home/ninja/dir2/file6: permission denied

DUPS: /home/ninja/dir1/file2 /home/ninja/dir2/file4
SH:mkdir -p 'finddups_4852402110976356155//home/ninja/dir2'
SH:mv '/home/ninja/dir2/file4' 'finddups_4852402110976356155//home/ninja/dir2/file4'

DUPS: /home/ninja/dir1/file2 /home/ninja/dir1/file1
SH:mkdir -p 'finddups_4852402110976356155//home/ninja/dir1'
SH:mv '/home/ninja/dir1/file1' 'finddups_4852402110976356155//home/ninja/dir1/file1'

SUMUP: Wasted space: 0.959 MB
```

 * Every line marked wiht 'DUPS:' gives you duplicate pair of files
 * Every line marked with 'SH:' gives you bash command that you might use to move the file in temp directory.
 * Every line marked with 'ERROR:' tells you about a proble with specific file
 * At the end there should be a line merked with 'SUMUP:', which gives you the wasted space


**To find total duplicate files:**

```./finddups /etc/ /tmp/ | egrep "^DUPS:" | wc -l```

**To extract the bash commands for execution:**

```./finddups /etc/ /tmp/| egrep "^SH:" |  awk -F "SH:" '{print $2}'```
