// mhinitdb provides a command that initializes a totally empty, but created,
// postgres database with a schema found in data/initial-pg.sql, and data from
// the neighboring Spells Compendium. Note that SSL is disabled, and the purpose
// here is only to save us from manually entering in all that juicy spell data
// that has already been compiled for us in that XML file.
package main
