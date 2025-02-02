package level

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/zeppelinmc/zeppelin/nbt"
)

// A game rule is a string containing an integer or a boolean
type GameRule string

func (rule GameRule) Boolean() (bool, error) {
	return strconv.ParseBool(string(rule))
}

func (rule GameRule) Integer() (int, error) {
	return strconv.Atoi(string(rule))
}

// A date represented by the amount of milliseconds since January 1st, 1970
type UnixMilliTimestamp int64

func (stamp UnixMilliTimestamp) Time() time.Time {
	return time.UnixMilli(int64(stamp))
}

type DimensionGenerationSettings struct {
	Generator struct {
		BiomeSource struct {
			Preset string `nbt:"preset,omitempty"`
			Type   string `nbt:"type"`
		} `nbt:"biome_source"`
		//Settings string `nbt:"settings"` // Can be both string and compound, skip for now
		Type string `nbt:"type"`
	} `nbt:"generator"`
	Type string `nbt:"type"`
}

type Level struct {
	Data struct {
		BorderCenterX        float64
		BorderCenterZ        float64
		BorderDamagePerBlock float64
		BorderSafeZone       float64
		BorderSizeLerpTarget float64
		BorderSizeLerpTime   int64
		BorderWarningBlocks  float64
		BorderWarningTime    float64

		DataPacks struct {
			Disabled []string
			Enabled  []string
		}

		DataVersion      int32
		DayTime          int64
		Difficulty       byte
		DifficultyLocked bool
		DragonFight      struct {
			DragonKilled       bool
			Gateways           []int32
			NeedsStateScanning bool
			PreviouslyKilled   byte
		}

		GameRules    map[string]GameRule
		GameType     GameMode
		LastPlayed   UnixMilliTimestamp
		LevelName    string
		ServerBrands []string

		SpawnAngle             float32
		SpawnX, SpawnY, SpawnZ int32

		Time    int64 // time since the world has started in ticks
		Version struct {
			Id       int32
			Name     string
			Series   string
			Snapshot int8
		}

		WanderingTraderId          []int32
		WanderingTraderSpawnChance int32
		WanderingTraderSpawnDelay  int32
		WasModded                  bool

		WorldGenSettings struct {
			BonusChest       bool                                   `nbt:"bonus_chest"`
			Dimensions       map[string]DimensionGenerationSettings `nbt:"dimensions"`
			GenerateFeatures bool                                   `nbt:"generate_features"`
			Seed             Seed                                   `nbt:"seed"`
		}

		AllowCommands    bool  `nbt:"allowCommands"`
		ClearWeatherTime int32 `nbt:"clearWeatherTime"`
		Hardcore         bool  `nbt:"hardcore"`
		Initialized      bool  `nbt:"initialized"`

		RainTime    int32 `nbt:"rainTime"`
		Raining     bool  `nbt:"raining"`
		Thundertime int32 `nbt:"thunderTime"`
		Thundering  bool  `nbt:"thundering"`
		VersionInt  int32 `nbt:"version"`
	}

	// the base path of the world
	basePath string `nbt:"-"`
}

// worldPath is the base path of the world
func LoadWorldLevel(worldPath string) (Level, error) {
	var level Level
	file, err := os.Open(worldPath + "/level.dat")
	if err != nil {
		return level, err
	}
	rd, err := gzip.NewReader(file)
	if err != nil {
		return level, err
	}

	var buf, _ = io.ReadAll(rd)

	rd.Close()
	file.Close()

	_, err = nbt.NewDecoder(bytes.NewReader(buf)).Decode(&level)
	level.basePath = worldPath

	return level, err
}

type Seed int64

// First 8 bytes of the SHA-256 hash of the world's seed. Used client side for biome noise
func (s Seed) HashedSeed() int64 {
	hash := sha256.Sum256(binary.BigEndian.AppendUint64(nil, uint64(s)))

	return int64(binary.BigEndian.Uint64(hash[:8]))
}
