filesync
====

Synchronizes a collection of files from a server for one or many clients across different platforms. You can use any operating system supported by go compiler for server and client.

Server and clients can be in different networks. Yes, you can sync over the internet.

## How it works?

Just complete the `conf.json` file and run the server. Do the same thing with the client.

The server read periodicaly the files list and build a sync mask. The client request periodicaly the sync mask and compare with local sync mask. If different datetime are found the files are queued for download.

Communication between server and clients is in JSON format (what else?).

## Config file

Both server and client use the same file format

     {
          "ip":"",
          "port":"8080",
          "path":"files",
          "synctime":"2"
     }

## Description

     "ip":"",
     "port":"8080",

Address and port for listening on server. No need to change `ip` but the port in a convenient value. For clients this is the address and port of the server.

      "path":"files",

For server and clients this is the files location. Not necessary to be the same on server and clients. Any subfolders will be synchronized.

     "synctime":"2"

Sync time represent the period measured in seconds until the next sincronization. Server and clients can have different values. For example the server can sync files at 10s and clients can request sync mask at 60s. Use proper values when you deal with big data to avoid overloading the machines.

Note that smaller sync time value will not override a sync in progress.

**Known issue:** Big number of files (tens of thousands) can overload the processor by numbering. So, keep the number of files in a reasonable quantity.