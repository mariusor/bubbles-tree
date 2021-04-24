package tree
const staticTree = `
Documents/
├── 70-smartcard.rules
├── 70-u2f.rules
├── 70-wifi-powersave.rules
├── AP.postman_collection.json
├── bugs
│   ├── blackmesa
│   │   ├── crash.png
│   │   └── dump.txt
│   └── exherbo
│       └── sway
│           └── install.log
├── cv
│   ├── coverletter-soundcloud.txt
│   ├── coverletter-zatoo.txt
│   ├── en
│   │   ├── additional[en].tex
│   │   ├── education[en].tex
│   │   ├── experience[en].tex
│   │   └── skills[en].tex
├── ebooks
│   ├── Delve
│   │   ├── Delve - 001: Woodland.mobi
│   │   ├── Delve - 002: One on One.mobi
│   │   ├── Delve - 003: Pothole.mobi
│   │   ├── Delve - 004: Statistics.mobi
│   │   ├── Delve - 005: Alone.mobi
├── expenses
│   └── 2020-08.md
├── Light_over_Darkness.xcf
├── My Games
│   └── The Vanishing of Ethan Carter
│       ├── AstronautsGame
│       │   ├── Cloud
│       │   │   └── CloudStorage.ini
│       │   ├── Config
│       │   │   ├── AstronautsConfig.inb
│       │   │   ├── AstronautsEngine.ini
│       │   │   ├── AstronautsGame.ini
│       │   │   ├── AstronautsInput.ini
│       │   │   ├── AstronautsLightmass.ini
│       │   │   ├── AstronautsSystemSettings.ini
│       │   │   └── AstronautsUI.ini
│       │   ├── Logs
│       │   │   ├── Launch-backup-2017.09.13-20.14.42.log
│       │   │   └── Launch.log
│       │   └── SaveData
│       │       └── SaveGame_0.sav
│       ├── Binaries
│       │   └── Win64
│       └── Engine
│           └── Config
│               └── ConsoleVariables.ini
├── ragel-guide-6.9.pdf
├── rules-for-polite-online-discourse.md
├── stories
│   ├── activitypub-c2s.md
│   ├── bulk_amorphous_metal.pdf
│   ├── butterflies_from_leonard.md
│   ├── coverlet.md
│   ├── gs
│   │   ├── chapter-eight
│   │   │   └── main.md
│   │   ├── chapter-eighteen
│   │   │   └── main.md
│   │   ├── chapter-eleven
│   │   │   └── main.md
│   │   ├── chapter-fifteen
│   │   │   └── main.md
│   │   ├── chapter-five
│   │   │   └── main.md
│   │   ├── chapter-four
│   │   │   ├── journey-start.md
│   │   │   └── main.md
│   │   ├── chapter-fourteen
│   │   │   └── main.md
│   │   ├── chapter-nine
│   │   │   └── main.md
│   │   ├── chapter-nineteen
│   │   │   └── main.md
│   │   ├── chapter-one
│   │   │   ├── dream.md
│   │   │   ├── main.md
│   │   │   ├── village-description.md
│   │   │   └── wake-up.md
│   │   ├── chapter-seven
│   │   │   └── main.md
│   │   ├── chapter-seventeen
│   │   │   └── main.md
│   │   ├── chapter-six
│   │   │   └── main.md
│   │   ├── chapter-sixteen
│   │   │   └── main.md
│   │   ├── chapter-ten
│   │   │   └── main.md
│   │   ├── chapter-thirteen
│   │   │   └── main.md
│   │   ├── chapter-three
│   │   │   └── main.md
│   │   ├── chapter-twelve
│   │   │   └── main.md
│   │   ├── chapter-two
│   │   │   └── main.md
│   │   ├── chapter-zero
│   │   │   ├── a-new-again.md
│   │   │   ├── commet.md
│   │   │   ├── disaster.md
│   │   │   ├── dream.md
│   │   │   ├── edit.md
│   │   │   ├── empty-valley.md
│   │   │   ├── i-am-not-really-here.md
│   │   │   ├── main.md
│   │   │   ├── memory_1.md
│   │   │   ├── old-man.md
│   │   │   ├── schema.md
│   │   │   └── snow.md
│   │   ├── excerpts
│   │   │   ├── below.md
│   │   │   ├── butterflies.md
│   │   │   ├── elk.md
│   │   │   ├── fairy-songs.md
│   │   │   ├── mother.md
│   │   │   ├── nordic-funeral-inscription.md
│   │   │   ├── one-hand.md
│   │   │   ├── strangers.md
│   │   │   ├── the-big-fizz.md
│   │   │   ├── the-wave.md
│   │   │   └── traveller.md
│   │   ├── external-references
│   │   │   ├── archaic-words.md
│   │   │   ├── arctic-dreams.md
│   │   │   ├── characters
│   │   │   │   ├── dogs.md
│   │   │   │   ├── families.md
│   │   │   │   ├── father.md
│   │   │   │   ├── guul
│   │   │   │   │   └── story.md
│   │   │   │   ├── jarl
│   │   │   │   │   └── dialogue-jarl.md
│   │   │   │   ├── little-brother.md
│   │   │   │   ├── main-character
│   │   │   │   │   ├── backstory-from-father.md
│   │   │   │   │   └── options.md
│   │   │   │   ├── main.md
│   │   │   │   └── mother.md
│   │   │   ├── main.md
│   │   │   ├── plagues.md
│   │   │   ├── religion
│   │   │   │   └── main.md
│   │   │   ├── setting
│   │   │   │   └── main.md
│   │   │   ├── snippets.md
│   │   │   └── snow-racing.md
│   │   ├── full.md
│   │   ├── gss.msk
│   │   ├── LICENSE
│   │   ├── meta.yml
│   │   └── toc.md
│   ├── hole.md
│   ├── LiquidMetal_01.pdf
│   ├── spk
│   │   ├── excerpts
│   │   │   ├── begining.md
│   │   │   └── snow.md
│   │   ├── external-references
│   │   │   ├── main.md
│   │   │   └── snippets.md
│   │   ├── LICENSE
│   │   ├── meta.yml
│   │   └── toc.md
│   ├── the-missing.md
│   └── the-unlikeness.md
├── tax.md
├── tuxedo-laptop-20200718.pdf
└── work
`
