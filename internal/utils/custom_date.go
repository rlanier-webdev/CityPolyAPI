// utils/customdate.go
package utils

import (
    "database/sql/driver"
    "fmt"
    "time"
)

type CustomDate struct {
    time.Time
}

func (cd *CustomDate) UnmarshalJSON(b []byte) error {
    t, err := time.Parse(`"2006-01-02"`, string(b))
    if err != nil {
        return err
    }
    cd.Time = t
    return nil
}

func (cd CustomDate) MarshalJSON() ([]byte, error) {
    return []byte(fmt.Sprintf(`"%s"`, cd.Time.Format("2006-01-02"))), nil
}

func (cd CustomDate) String() string {
    return cd.Time.Format("2006-01-02")
}

// Implement the Valuer interface
func (cd CustomDate) Value() (driver.Value, error) {
    return cd.Time, nil
}

// Implement the Scanner interface
func (cd *CustomDate) Scan(value interface{}) error {
    if value == nil {
        *cd = CustomDate{Time: time.Time{}}
        return nil
    }

    switch v := value.(type) {
    case time.Time:
        *cd = CustomDate{Time: v}
        return nil
    case string:
        t, err := time.Parse("2006-01-02", v)
        if err != nil {
            return err
        }
        *cd = CustomDate{Time: t}
        return nil
    default:
        return fmt.Errorf("cannot scan type %T into CustomDate: %v", value, value)
    }
}
