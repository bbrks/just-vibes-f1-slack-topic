package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// F1 API base URL
const BaseURL = "https://f1api.dev/api"

// Driver Championship response
type DriverChampionshipResponse struct {
	API                 string           `json:"api"`
	URL                 string           `json:"url"`
	Limit               int              `json:"limit"`
	Offset              int              `json:"offset"`
	Total               int              `json:"total"`
	Season              int              `json:"season"`
	ChampionshipID      string           `json:"championshipId"`
	DriversChampionship []DriverStanding `json:"drivers_championship"`
}

// Driver Standing represents a driver in the championship
type DriverStanding struct {
	ClassificationID int     `json:"classificationId"`
	DriverID         string  `json:"driverId"`
	TeamID           string  `json:"teamId"`
	Points           float64 `json:"points"`
	Position         int     `json:"position"`
	Wins             int     `json:"wins"`
	Driver           Driver  `json:"driver"`
	Team             Team    `json:"team"`
}

// Constructor Championship response
type ConstructorChampionshipResponse struct {
	API                      string         `json:"api"`
	URL                      string         `json:"url"`
	Limit                    int            `json:"limit"`
	Offset                   int            `json:"offset"`
	Total                    int            `json:"total"`
	Season                   int            `json:"season"`
	ChampionshipID           string         `json:"championshipId"`
	ConstructorsChampionship []TeamStanding `json:"constructors_championship"`
}

// Team Standing represents a team in the championship
type TeamStanding struct {
	ClassificationID int     `json:"classificationId"`
	TeamID           string  `json:"teamId"`
	Points           float64 `json:"points"`
	Position         int     `json:"position"`
	Wins             int     `json:"wins"`
	Team             Team    `json:"team"`
}

// Next Race response
type NextRaceResponse struct {
	API          string       `json:"api"`
	URL          string       `json:"url"`
	Total        int          `json:"total"`
	Season       int          `json:"season"`
	Round        int          `json:"round"`
	Championship Championship `json:"championship"`
	Race         []Race       `json:"race"`
}

// Championship represents an F1 championship
type Championship struct {
	ChampionshipID   string `json:"championshipId"`
	ChampionshipName string `json:"championshipName"`
	URL              string `json:"url"`
	Year             int    `json:"year"`
}

// Race represents a Formula 1 race
type Race struct {
	RaceID         string   `json:"raceId"`
	ChampionshipID string   `json:"championshipId"`
	RaceName       string   `json:"raceName"`
	Schedule       Schedule `json:"schedule"`
	Circuit        Circuit  `json:"circuit"`
	Country        string   `json:"country"`
	Sprint         bool     `json:"sprint"`
}

// Schedule represents the schedule for a race weekend
type Schedule struct {
	Race  TimeInfo `json:"race"`
	Qualy TimeInfo `json:"qualy"`
}

// TimeInfo contains date and time information
type TimeInfo struct {
	Date string `json:"date"`
	Time string `json:"time"`
}

// Circuit represents a Formula 1 circuit
type Circuit struct {
	CircuitID   string  `json:"circuitId"`
	CircuitName string  `json:"circuitName"`
	Length      float64 `json:"length"`
	Laps        int     `json:"laps"`
	Laprecord   string  `json:"lapRecord,omitempty"`
}

// Driver represents a Formula 1 driver
type Driver struct {
	Name        string `json:"name"`
	Surname     string `json:"surname"`
	Nationality string `json:"nationality"`
	Birthday    string `json:"birthday"`
	Number      int    `json:"number"`
	ShortName   string `json:"shortName"`
	URL         string `json:"url"`
}

// Team represents a Formula 1 team
type Team struct {
	TeamID                    string `json:"teamId,omitempty"`
	TeamName                  string `json:"teamName"`
	Country                   string `json:"country"`
	FirstAppearance           int    `json:"firstAppareance,omitempty"`
	ConstructorsChampionships int    `json:"constructorsChampionships,omitempty"`
	DriversChampionships      int    `json:"driversChampionships,omitempty"`
	URL                       string `json:"url"`
}

// Error response when no data is found
type ErrorResponse struct {
	API     string `json:"api"`
	URL     string `json:"url"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

// Map team IDs to team emojis and abbreviations
var teamEmojis = map[string]struct {
	emoji string
	abbr  string
}{
	"mclaren":      {":m1::f1tl:", "MCL"},
	"mercedes":     {":f1tm:", "MER"},
	"red_bull":     {":f1tr:", "RBR"},
	"ferrari":      {":f1tf:", "FER"},
	"aston_martin": {":f1ta:", "AST"},
	"williams":     {":f1tw:", "WIL"},
	"alpine":       {":f1ta:", "ALP"},
	"haas":         {":f1th:", "HAA"},
	"rb":           {":f1tb:", "RB"},
	"sauber":       {":f1ts:", "SAU"},
}

// Map driver IDs to driver emojis
var driverEmojis = map[string]string{
	"norris":         ":f1ln:",
	"max_verstappen": ":f1mv:",
	"russell":        ":f1gr:",
	"piastri":        ":f1op:",
	"hamilton":       ":f1lh:",
	"leclerc":        ":f1cl:",
	"sainz":          ":f1cs:",
	"alonso":         ":f1fa:",
	"stroll":         ":f1ls:",
	"albon":          ":f1aa:",
	"tsunoda":        ":f1yt:",
	"hulkenberg":     ":f1nh:",
	"ocon":           ":f1eo:",
	"gasly":          ":f1pg:",
}

// Map country names to flag emojis
var countryFlags = map[string]string{
	"Great Britain": ":gb:",
	"Netherlands":   ":flag-nl:",
	"Australia":     ":flag-au:",
	"Monaco":        ":flag-mc:",
	"Spain":         ":flag-es:",
	"Mexico":        ":flag-mx:",
	"Canada":        ":flag-ca:",
	"Japan":         ":flag-jp:",
	"France":        ":flag-fr:",
	"Thailand":      ":flag-th:",
	"China":         ":flag-cn:",
	"United States": ":flag-us:",
	"Italy":         ":flag-it:",
	"Germany":       ":flag-de:",
}

// Map countries to their 2-letter codes for flag emojis
var countryTwoLetterCodes = map[string]string{
	"China":                "cn",
	"Japan":                "jp",
	"Great Britain":        "gb",
	"United Kingdom":       "gb",
	"United States":        "us",
	"USA":                  "us",
	"Australia":            "au",
	"Netherlands":          "nl",
	"Italy":                "it",
	"France":               "fr",
	"Germany":              "de",
	"Spain":                "es",
	"Monaco":               "mc",
	"Canada":               "ca",
	"Mexico":               "mx",
	"Brazil":               "br",
	"Austria":              "at",
	"Belgium":              "be",
	"Hungary":              "hu",
	"Singapore":            "sg",
	"Russia":               "ru",
	"Azerbaijan":           "az",
	"Bahrain":              "bh",
	"United Arab Emirates": "ae",
	"Abu Dhabi":            "ae",
	"Qatar":                "qa",
	"Saudi Arabia":         "sa",
	"Thailand":             "th",
	"Switzerland":          "ch",
}

// fetchNextRace gets the next race in the calendar and its round number
func fetchNextRace() (*Race, int, error) {
	url := fmt.Sprintf("%s/current/next", BaseURL)

	log.Printf("Fetching next race data from: %s", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, 0, fmt.Errorf("error fetching next race: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("error reading response: %v", err)
	}

	// Log a truncated version of the response for debugging
	truncLen := min(len(body), 500)
	log.Printf("Next race API response (truncated): %s", string(body[:truncLen]))

	// Check if we got an error response
	var errorResp ErrorResponse
	if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Status >= 400 {
		return nil, 0, fmt.Errorf("no data found: %s", errorResp.Message)
	}

	var nextRaceResp NextRaceResponse
	if err := json.Unmarshal(body, &nextRaceResp); err != nil {
		return nil, 0, fmt.Errorf("error unmarshaling next race data: %v", err)
	}

	if len(nextRaceResp.Race) == 0 {
		return nil, 0, fmt.Errorf("no upcoming races found")
	}

	return &nextRaceResp.Race[0], nextRaceResp.Round, nil
}

// fetchDriverStandings gets the current driver championship standings
func fetchDriverStandings() ([]DriverStanding, error) {
	url := fmt.Sprintf("%s/current/drivers-championship", BaseURL)

	log.Printf("Fetching driver standings from: %s", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching driver standings: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	// Log a truncated version of the response for debugging
	truncLen := min(len(body), 500)
	log.Printf("Driver standings API response (truncated): %s", string(body[:truncLen]))

	// Check if we got an error response
	var errorResp ErrorResponse
	if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Status >= 400 {
		return nil, fmt.Errorf("no data found: %s", errorResp.Message)
	}

	var driverResp DriverChampionshipResponse
	if err := json.Unmarshal(body, &driverResp); err != nil {
		return nil, fmt.Errorf("error unmarshaling driver data: %v", err)
	}

	return driverResp.DriversChampionship, nil
}

// fetchTeamStandings gets the current constructor championship standings
func fetchTeamStandings() ([]TeamStanding, error) {
	url := fmt.Sprintf("%s/current/constructors-championship", BaseURL)

	log.Printf("Fetching team standings from: %s", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching team standings: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	// Log a truncated version of the response for debugging
	truncLen := min(len(body), 500)
	log.Printf("Team standings API response (truncated): %s", string(body[:truncLen]))

	// Check if we got an error response
	var errorResp ErrorResponse
	if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Status >= 400 {
		return nil, fmt.Errorf("no data found: %s", errorResp.Message)
	}

	var teamResp ConstructorChampionshipResponse
	if err := json.Unmarshal(body, &teamResp); err != nil {
		return nil, fmt.Errorf("error unmarshaling team data: %v", err)
	}

	return teamResp.ConstructorsChampionship, nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Topic builds the F1 information string
func Topic() string {
	var sb strings.Builder

	// Get current season year
	currentYear := time.Now().Year()
	sb.WriteString(fmt.Sprintf("F1 Data for %d\n\n", currentYear))

	// Get and display next race
	nextRace, round, err := fetchNextRace()
	if err != nil {
		log.Printf("Error getting next race: %v", err)
		sb.WriteString(fmt.Sprintf("Next race: %v\n\n", err))
	} else {
		raceDate, err := time.Parse("2006-01-02", nextRace.Schedule.Race.Date)
		if err != nil {
			log.Printf("Error parsing race date: %v", err)
			sb.WriteString("Next Race: Unknown (date parsing error)\n\n")
		} else {
			raceTime, timeErr := time.Parse("15:04:05Z", nextRace.Schedule.Race.Time)
			timeStr := ""
			if timeErr == nil {
				timeStr = fmt.Sprintf(" at %s", raceTime.Format("15:04 MST"))
			}

			sb.WriteString(fmt.Sprintf("Next Race: %s (Round %d)\n", nextRace.RaceName, round))
			sb.WriteString(fmt.Sprintf("Circuit: %s\n", nextRace.Circuit.CircuitName))
			sb.WriteString(fmt.Sprintf("Date: %s%s\n", raceDate.Format("January 2, 2006"), timeStr))
			sb.WriteString(fmt.Sprintf("Country: %s\n\n", nextRace.Country))
		}
	}

	// Get and display driver standings
	drivers, err := fetchDriverStandings()
	if err != nil {
		log.Printf("Error fetching driver standings: %v", err)
		sb.WriteString(fmt.Sprintf("Driver standings error: %v\n\n", err))
	} else {
		sb.WriteString("Driver Standings:\n")
		for i := 0; i < 3 && i < len(drivers); i++ {
			driver := drivers[i]
			sb.WriteString(fmt.Sprintf("%d. %s %s (%s) - %.1f points\n",
				driver.Position, driver.Driver.Name, driver.Driver.Surname, driver.Team.TeamName, driver.Points))
		}
		sb.WriteString("\n")
	}

	// Get and display constructor/team standings
	teams, err := fetchTeamStandings()
	if err != nil {
		log.Printf("Error fetching team standings: %v", err)
		sb.WriteString(fmt.Sprintf("Constructor standings error: %v\n", err))
	} else {
		sb.WriteString("Constructor Standings:\n")
		for i := 0; i < 3 && i < len(teams); i++ {
			team := teams[i]
			sb.WriteString(fmt.Sprintf("%d. %s - %.1f points\n",
				team.Position, team.Team.TeamName, team.Points))
		}
	}

	return sb.String()
}

// SlackTopic builds a compact Slack topic with emojis for F1 information
func SlackTopic() string {
	var sb strings.Builder

	// Get current season year
	currentYear := time.Now().Year()
	totalRaces := 24 // Hardcoded for now, could be retrieved from API

	// Start with F1 emoji and year
	sb.WriteString(fmt.Sprintf(":f1: %d ", currentYear))

	// Get next race
	nextRace, round, err := fetchNextRace()
	if err != nil {
		log.Printf("Error getting next race: %v", err)
		sb.WriteString("Next: No upcoming races // ")
	} else {
		// Parse race date
		raceDate, err := time.Parse("2006-01-02", nextRace.Schedule.Race.Date)
		if err != nil {
			log.Printf("Error parsing race date: %v", err)
			sb.WriteString(fmt.Sprintf("Next: R%d/%d %s // ", round, totalRaces, extractRaceName(nextRace.RaceName)))
		} else {
			// Get country code for flag emoji
			countryCode := "unknown"

			// First try to determine country from race name if the Country field is empty
			raceName := extractRaceName(nextRace.RaceName)
			if raceName == "Japan" || strings.Contains(strings.ToLower(nextRace.RaceName), "japanese") {
				countryCode = "jp"
			} else if raceName == "China" || strings.Contains(strings.ToLower(nextRace.RaceName), "chinese") {
				countryCode = "cn"
			} else if len(nextRace.Country) > 0 {
				// If Country field is set, try to get the code from our map
				if code, exists := countryTwoLetterCodes[nextRace.Country]; exists {
					countryCode = code
				} else {
					// Default to first two letters of country name, lowercase
					if len(nextRace.Country) >= 2 {
						countryCode = strings.ToLower(nextRace.Country[0:2])
					}
				}
			}

			// Calculate race weekend dates (Friday-Sunday)
			raceWeekendStart := raceDate.AddDate(0, 0, -2) // Friday is typically 2 days before race day (Sunday)

			// Format as "Next: R[round]/24 [race] :flag-xx: (Mon DD-DD)"
			sb.WriteString(fmt.Sprintf("Next: R%d/%d %s :flag-%s: (%s %d-%d) // ",
				round,
				totalRaces,
				extractRaceName(nextRace.RaceName),
				countryCode,
				raceWeekendStart.Format("Jan"),
				raceWeekendStart.Day(),
				raceDate.Day()))
		}
	}

	// Get driver standings
	drivers, err := fetchDriverStandings()
	if err != nil {
		log.Printf("Error fetching driver standings: %v", err)
		sb.WriteString("Standings: No data // ")
	} else {
		sb.WriteString("Standings: ")

		// Add top 3 drivers
		for i := 0; i < 3 && i < len(drivers); i++ {
			driver := drivers[i]

			// Get driver emoji
			driverEmoji := ""
			if emoji, exists := driverEmojis[driver.DriverID]; exists {
				driverEmoji = emoji
			}

			// Get flag emoji
			flagEmoji := ":flag-xx:"
			if flag, exists := countryFlags[driver.Driver.Nationality]; exists {
				flagEmoji = flag
			}

			// Format as "[driverEmoji]ABBR flagEmoji (points)"
			sb.WriteString(fmt.Sprintf("%s%s %s (%.0f)",
				driverEmoji,
				driver.Driver.ShortName,
				flagEmoji,
				driver.Points))

			// Add comma if not the last driver
			if i < 2 && i+1 < len(drivers) {
				sb.WriteString(", ")
			}
		}

		// Separator between drivers and constructors
		sb.WriteString("; ")
	}

	// Get constructor standings
	teams, err := fetchTeamStandings()
	if err != nil {
		log.Printf("Error fetching team standings: %v", err)
		sb.WriteString("No constructor data // ")
	} else {
		// Add top 3 constructors
		for i := 0; i < 3 && i < len(teams); i++ {
			team := teams[i]

			// Get team emoji and abbreviation
			teamInfo := teamEmojis["unknown"]
			if info, exists := teamEmojis[team.TeamID]; exists {
				teamInfo = info
			}

			// Format as "teamEmoji ABBR (points)"
			sb.WriteString(fmt.Sprintf("%s%s (%.0f)",
				teamInfo.emoji,
				teamInfo.abbr,
				team.Points))

			// Add comma if not the last team
			if i < 2 && i+1 < len(teams) {
				sb.WriteString(", ")
			}
		}
	}

	// Add fantasy code
	sb.WriteString(" // Fantasy: `C14SOD0WQ01`")

	// Get the final topic string
	topic := sb.String()

	// Check if the topic exceeds the 250 character limit
	// Slack counts each character, including emoji codes (e.g., ":flag-jp:" is 9 characters)
	if len(topic) > 250 {
		log.Printf("WARNING: Slack topic exceeds 250 character limit (%d characters)", len(topic))
		return fmt.Sprintf("ERROR: Slack topic exceeds 250 character limit (%d characters)", len(topic))
	}

	return topic
}

// extractRaceName extracts the main part of the race name (e.g., "Lenovo Japanese Grand Prix 2025" -> "Japan")
func extractRaceName(fullName string) string {
	// This is a simple implementation - could be more sophisticated
	if strings.Contains(fullName, "Japanese") {
		return "Japan"
	} else if strings.Contains(fullName, "Chinese") {
		return "China"
	} else if strings.Contains(fullName, "Monaco") {
		return "Monaco"
	} else if strings.Contains(fullName, "British") {
		return "Britain"
	} else if strings.Contains(fullName, "Italian") {
		return "Italy"
	} else if strings.Contains(fullName, "Spanish") {
		return "Spain"
	} else if strings.Contains(fullName, "Australian") {
		return "Australia"
	} else if strings.Contains(fullName, "Hungarian") {
		return "Hungary"
	} else if strings.Contains(fullName, "Belgian") {
		return "Belgium"
	} else if strings.Contains(fullName, "Dutch") {
		return "Netherlands"
	} else if strings.Contains(fullName, "United States") || strings.Contains(fullName, "USA") {
		return "USA"
	} else if strings.Contains(fullName, "Mexico") || strings.Contains(fullName, "Mexican") {
		return "Mexico"
	} else if strings.Contains(fullName, "Brazil") || strings.Contains(fullName, "Brazilian") || strings.Contains(fullName, "São Paulo") {
		return "Brazil"
	} else if strings.Contains(fullName, "Abu Dhabi") {
		return "Abu Dhabi"
	} else if strings.Contains(fullName, "Qatar") {
		return "Qatar"
	} else if strings.Contains(fullName, "Singapore") {
		return "Singapore"
	} else if strings.Contains(fullName, "Saudi") || strings.Contains(fullName, "Jeddah") {
		return "Saudi Arabia"
	} else if strings.Contains(fullName, "Bahrain") {
		return "Bahrain"
	} else if strings.Contains(fullName, "Miami") {
		return "Miami"
	} else if strings.Contains(fullName, "Las Vegas") {
		return "Las Vegas"
	} else if strings.Contains(fullName, "Canada") || strings.Contains(fullName, "Canadian") || strings.Contains(fullName, "Montreal") {
		return "Canada"
	} else if strings.Contains(fullName, "Azerbaijan") || strings.Contains(fullName, "Baku") {
		return "Azerbaijan"
	}

	// Return a shortened version if no specific match
	parts := strings.Split(fullName, " ")
	if len(parts) > 2 {
		return strings.Join(parts[1:len(parts)-2], " ") // Skip first word (usually a sponsor) and last two (usually "Grand Prix YYYY")
	}
	return fullName
}

func main() {
	// Define command-line flags
	detailed := flag.Bool("detailed", true, "Show detailed output (default)")
	slackFormat := flag.Bool("slack", false, "Show Slack topic format")
	quiet := flag.Bool("quiet", false, "Suppress log messages")
	flag.Parse()

	// If -quiet flag is set, disable logging
	if *quiet {
		log.SetOutput(io.Discard)
	}

	// Choose output format based on flags
	if *slackFormat {
		*detailed = false
		topic := SlackTopic()
		fmt.Println(topic)

		// Check if topic contains an error about exceeding character limit
		if strings.HasPrefix(topic, "ERROR:") {
			os.Exit(1)
		}
	} else if *detailed {
		fmt.Println(Topic())
	} else {
		// Default to detailed if no format is specified
		fmt.Println(Topic())
	}
}
