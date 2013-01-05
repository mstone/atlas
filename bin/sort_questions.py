#!/usr/bin/env python
import sys
import json

def question_cmp(x, y):
    a = cmp(x['GroupKey'], y['GroupKey'])
    if a != 0:
        return a
    else:
        return cmp(x['SortKey'], y['SortKey'])

def main():
    infile = open(sys.argv[1])
    outfile = open(sys.argv[2], "w")

    root = json.load(infile)

    for p in root["Profiles"]:
        p["Questions"].sort(cmp=question_cmp)

    json.dump(root, outfile, indent=4)

if __name__ == "__main__":
    main()
