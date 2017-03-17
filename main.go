package main

import (
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/murder-hobos/murder-hobos/db/initDb"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	user, passwd, host, port, dbname string
	dropEverythingAndInitialize      string
	xmlBytes                         []byte
	help                             bool
)

const (
	xmlFilePath = "data/Spells Compendium 1.2.1.xml"
	sqlFilePath = "data/initial-pg.sql"
)

func init() {
	flag.StringVar(&user, "U", os.Getenv("USER"), "Database user name")
	flag.StringVar(&passwd, "W", "", "Database password")
	flag.StringVar(&host, "h", "localhost", "Host name")
	flag.StringVar(&port, "p", "5432", "Port number")
	flag.StringVar(&dbname, "d", "", "Database name (required)")
	flag.BoolVar(&help, "help", false, "Displays this help")

	// Retrieve sql/xml info from bindata bundled with this executable
	sqlBytes, err := initDb.Asset(sqlFilePath)
	if err != nil {
		log.Fatalln(err)
	}
	dropEverythingAndInitialize = string(sqlBytes)

	xmlBytes, err = initDb.Asset(xmlFilePath)
	if err != nil {
		log.Fatalln(err)
	}
}

func confirm() bool {
	color.Set(color.FgRed)
	fmt.Print("WARNING: ")
	color.Unset()
	fmt.Println("All data in database will be erased and replaced with starting data.")
	fmt.Println("That means user data too.")
	fmt.Print("Are you sure you want to continue? [y\\N] ")

	var resp string
	_, err := fmt.Scanln(&resp)
	if err != nil {
		if err.Error() == "unexpected newline" {
			os.Exit(0)
		} else {
			log.Fatalln(err)
		}
	}
	return resp == "Y" || resp == "y"
}

func main() {
	flag.Parse()

	if help {
		flag.Usage()
		os.Exit(1)
	}

	if dbname == "" {
		fmt.Println("Error: Database name is required")
		flag.Usage()
		os.Exit(1)
	}
	if !confirm() {
		os.Exit(1)
	}

	if passwd == "" {
		fmt.Print("Password: ")
		// Don't echo password out
		pass, err := terminal.ReadPassword(0)
		if err != nil {
			log.Fatalln("Fine. Don't enter a password. Bye.")
		}
		passwd = string(pass)
		fmt.Println()
	}

	b := &bytes.Buffer{}
	b.WriteString("user=")
	b.WriteString(user)
	b.WriteString(" dbname=")
	b.WriteString(dbname)
	b.WriteString(" password=")
	b.WriteString(passwd)
	b.WriteString(" host=")
	b.WriteString(host)
	b.WriteString(" port=")
	b.WriteString(port)
	b.WriteString(" sslmode=disable")
	db, err := sqlx.Connect("postgres", b.String())

	if err != nil {
		log.Fatalln(err)
	}

	if _, err := db.Exec(dropEverythingAndInitialize); err != nil {
		log.Fatalln(err)
	}

	// Have to be silly about this because range is a reserved word
	insertSpell, err := db.PrepareNamed(`
		INSERT INTO spell (name, level, school, cast_time, duration,
		"range", comp_verbal, comp_somatic, comp_material, material_desc,
        concentration, ritual, description, source_id)
		VALUES
		(:name, :level, :school, :cast_time, :duration, :range, :comp_verbal,
        :comp_somatic, :comp_material, :material_desc, :concentration, :ritual,
		:description, :source_id)
        RETURNING id;
	`)
	if err != nil {
		log.Fatalln(err)
	}

	insertClassSpells, err := db.Prepare(`
		INSERT INTO class_spells (spell_id, class_id) VALUES ($1, $2);
	`)
	if err != nil {
		log.Fatalln(err)
	}

	var c initDb.Compendium
	if err := xml.Unmarshal(xmlBytes, &c); err != nil {
		log.Fatalln(err)
	}

	// for each spell in our xml file
	for _, xmlSpell := range c.XMLSpells {
		s, err := xmlSpell.ToDbSpell()
		if err != nil {
			log.Fatalln("Error converting to db spell")
		}

		// Insert into spell table
		var spellID int
		err = insertSpell.QueryRowx(&s).Scan(&spellID)
		//result, err := insertSpell.Exec(&s)
		if err != nil {
			log.Fatalln(err)
		}

		// Insert into class_spells table
		if classes, ok := xmlSpell.ParseClasses(); ok {
			for _, class := range classes {
				if _, err := insertClassSpells.Exec(spellID, class.ID); err != nil {
					log.Fatalln(err)
				}
			}
		} else {
			log.Fatalf("Error parsing classes from %v\n", xmlSpell)
		}
	}
}
