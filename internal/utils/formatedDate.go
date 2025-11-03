package utils

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const layout = "2006-01-02 15:04:05"

type FormattedTime struct {
	time.Time
}

// MarshalJSON formatea al exportar a JSON
func (ft FormattedTime) MarshalJSON() ([]byte, error) {
	formatted := ft.Format(layout)
	return json.Marshal(formatted)
}

// UnmarshalJSON permite parsear desde JSON
func (ft *FormattedTime) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	if strings.TrimSpace(str) == "" {
		ft.Time = time.Time{}
		return nil
	}
	t, err := time.Parse(layout, str)
	if err != nil {
		return err
	}
	ft.Time = t
	return nil
}

// Scan permite a GORM leer el campo desde la DB
func (ft *FormattedTime) Scan(value interface{}) error {
	if value == nil {
		ft.Time = time.Time{}
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		ft.Time = v
	case []byte:
		t, err := time.Parse(layout, string(v))
		if err != nil {
			return err
		}
		ft.Time = t
	case string:
		t, err := time.Parse(layout, v)
		if err != nil {
			return err
		}
		ft.Time = t
	default:
		return fmt.Errorf("cannot scan type %T into FormattedTime", value)
	}
	return nil
}

// Value permite a GORM guardar el campo en la DB
func (ft FormattedTime) Value() (driver.Value, error) {
	return ft.Truncate(time.Second).Format(layout), nil
}

func (ft FormattedTime) String() string {
	return ft.Format(layout)
}

func NowFormatted() FormattedTime {
	return FormattedTime{Time: time.Now().Truncate(time.Second)}
}
