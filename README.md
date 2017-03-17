# mhinitdb 

### Murder Hobos Initialize Database

To install ```mhinitdb``` to your system:
```
go get -u github.com/murder-hobos/mhinitdb
```

This package provides a command ```mhinitdb``` that bootstraps an empty postgres database for use
with the Murder Hobos application. The database must already be created, and the command must be
supplied with a user that can create tables. The point of this package is to save us the tedium
of manually entering the data for all of those spells in the xml file that some people more dedicated
than us have already compiled.

A dump of the initial state created by running this should be much
more efficient for creating our initial state later.

***WARNING:*** Running this command wipes the tables. 

Usage:
```
mhinitdb -U username -d database-name  -W password -h hostname -p port
```
