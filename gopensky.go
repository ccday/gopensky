package gopensky

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const Root = "https://opensky-network.org/api"

type Api interface {
	Get(req *Request) (*Response, error)
}

type Request struct {
	// The time in seconds since epoch (Unix timestamp to retrieve states for. Current time will be used if omitted.
	Time int
	// One or more ICAO24 transponder addresses represented by a hex string (e.g. abc9f3). To filter multiple ICAO24
	// append the property once for each address. If omitted, the state vectors of all aircraft are returned.
	Icao24 []string

	Bbox *Bbox
}

type Bbox struct {
	// Lower bound for the latitude in decimal degrees.
	Lamin float64
	// Lower bound for the longitude in decimal degrees.
	Lomin float64
	// Upper bound for the latitude in decimal degrees.
	Lamax float64
	// Upper bound for the longitude in decimal degrees.
	Lomax float64
}

type Response struct {
	// The time which the state vectors in this response are associated with. All vectors represent the state of a vehicle
	// with the interval [time−1,time].
	Time int
	// The state vectors.
	States []*State
}

type State struct {
	// Unique ICAO 24-bit address of the transponder in hex string representation.
	Icao24 string
	// Callsign of the vehicle (8 chars). Can be null if no callsign has been received.
	Callsign string
	// Country name inferred from the ICAO 24-bit address.
	OriginCountry string
	// Unix timestamp (seconds) for the last position update. Can be null if no position report was received by OpenSky
	// within the past 15s.
	TimePosition int
	// Unix timestamp (seconds) for the last update in general. This field is updated for any new, valid message received
	// from the transponder.
	LastContact int
	// WGS-84 longitude in decimal degrees. Can be null.
	Longitude float64
	// WGS-84 latitude in decimal degrees. Can be null.
	Latitude float64
	// Barometric altitude in meters. Can be null.
	BaroAltitude float64
	// Boolean value which indicates if the position was retrieved from a surface position report.
	OnGround bool
	// Velocity over ground in m/s. Can be null.
	Velocity float64
	// True track in decimal degrees clockwise from north (north=0°). Can be null.
	TrueTrack float64
	// Vertical rate in m/s. A positive value indicates that the airplane is climbing, a negative value indicates that it
	// descends. Can be null.
	VerticalRate float64
	// IDs of the receivers which contributed to this state vector. Is null if no filtering for sensor was used in the
	// request.
	Sensors []int
	// Geometric altitude in meters. Can be null.
	GeoAltitude float64
	// The transponder code aka Squawk. Can be null.
	Squawk string
	// Whether flight status indicates special purpose indicator.
	Spi bool
	// Origin of this state’s position: 0 = ADS-B, 1 = ASTERIX, 2 = MLAT
	PositionSource int
}

type api struct {
	Http *http.Client
}

func New(httpClient *http.Client) Api {
	return &api{httpClient}
}

func (a *api) Get(req *Request) (*Response, error) {
	u := endpointFor("states", "all")
	if req != nil {
		u.RawQuery = serializeQueryParams(req)
	}

	res, err := a.Http.Get(u.String())
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("GET %s not OK: %s", u.String(), res.Status)
	}

	var raw map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
		return nil, err
	}
	states := deserializeStates(raw["states"].([]interface{}))

	return &Response{
		Time:   int(raw["time"].(float64)),
		States: states,
	}, nil
}

func endpointFor(path ...string) (u *url.URL) {
	u, _ = url.Parse(strings.Join(append([]string{Root}, path...), "/"))
	return
}

func formatDegrees(degrees float64) string {
	return fmt.Sprintf("%.4f", degrees)
}

func serializeQueryParams(req *Request) string {
	v := url.Values{}

	if req.Time != 0 {
		v.Set("time", strconv.Itoa(req.Time))
	}

	for _, icao24 := range req.Icao24 {
		v.Add("icao24", icao24)
	}

	if req.Bbox != nil {
		v.Set("lamin", formatDegrees(req.Bbox.Lamin))
		v.Set("lomin", formatDegrees(req.Bbox.Lomin))
		v.Set("lamax", formatDegrees(req.Bbox.Lamax))
		v.Set("lomax", formatDegrees(req.Bbox.Lomax))
	}

	return v.Encode()
}

func deserializeStates(rawStates []interface{}) []*State {
	states := make([]*State, 0)
	for _, rawState := range rawStates {
		states = append(states, deserializeState(rawState))
	}
	return states
}

func deserializeState(state interface{}) *State {
	vec := state.([]interface{})
	return &State{
		Icao24:         vec[0].(string),
		Callsign:       deserializeString(vec[1]),
		OriginCountry:  vec[2].(string),
		TimePosition:   int(deserializeFloat64(vec[3])),
		LastContact:    int(vec[4].(float64)),
		Longitude:      deserializeFloat64(vec[5]),
		Latitude:       deserializeFloat64(vec[6]),
		BaroAltitude:   deserializeFloat64(vec[7]),
		OnGround:       vec[8].(bool),
		Velocity:       deserializeFloat64(vec[9]),
		TrueTrack:      deserializeFloat64(vec[10]),
		VerticalRate:   deserializeFloat64(vec[11]),
		Sensors:        deserializeIntSlice(vec[12]),
		GeoAltitude:    deserializeFloat64(vec[13]),
		Squawk:         deserializeString(vec[14]),
		Spi:            vec[15].(bool),
		PositionSource: int(vec[16].(float64)),
	}
}

func deserializeIntSlice(slice interface{}) []int {
	if slice == nil {
		return []int{}
	}

	return slice.([]int)
}

func deserializeString(str interface{}) string {
	if str == nil {
		return ""
	}

	return str.(string)
}

func deserializeFloat64(f64 interface{}) float64 {
	if f64 == nil {
		return 0
	}

	return f64.(float64)
}
