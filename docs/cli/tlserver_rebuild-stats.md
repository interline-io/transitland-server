## tlserver rebuild-stats

Rebuild statistics for feeds or specific feed versions

### Synopsis

Rebuild statistics for feeds or specific feed versions



```
tlserver rebuild-stats [flags] [feeds...]
```

### Options

```
      --dburl string                       Database URL (default: $TL_DATABASE_URL)
      --fv-sha1 strings                    Feed version SHA1
      --fv-sha1-file string                Specify feed version IDs by SHA1 in file, one per line
      --fvid strings                       Rebuild stats for specific feed version ID
      --fvid-file string                   Specify feed version IDs in file, one per line; equivalent to multiple --fvid
  -h, --help                               help for rebuild-stats
      --storage string                     Storage destination; can be s3://... az://... or path to a directory
      --validation-report                  Save validation report
      --validation-report-storage string   Storage path for saving validation report JSON
      --workers int                        Worker threads (default 1)
```

### SEE ALSO

* [tlserver](tlserver.md)	 - 

###### Auto generated by spf13/cobra on 17-Aug-2024