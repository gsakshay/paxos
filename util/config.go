package util

import (
	"bufio"
	"flag"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	ProposerRole = "proposer"
	AcceptorRole = "acceptor"
	LearnerRole  = "learner"
)

type HostInfo struct {
	Hostname string
	Proposer []int64
	Acceptor []int64
	Learner  []int64
}

func ParseFlags() (string, string, float64) {
	hostfile := flag.String("h", "", "Path to the hostfile")
	proposerValue := flag.String("v", "", "Proposer value")
	timeDelay := flag.Float64("t", 0.0, "Time delay in seconds to wait before sending a proposal")

	// Parse command-line flags
	flag.Parse()

	return *hostfile, *proposerValue, *timeDelay
}

// ReadHostfile reads the hostfile and returns a map where keys are line numbers (ID) and values are HostInfo.
// Additionally, it returns a quorum map indicating which acceptors are associated with each proposer.
func ReadHostfile(fileName string) (map[int64]HostInfo, map[int64][]int64) {
	hostRoles := make(map[int64]HostInfo)
	quorumMap := make(map[int64][]int64)

	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("failed to open hostfile: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lineID int64 = 1
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			log.Fatalf("invalid format in hostfile: %s", line)
		}

		hostname := parts[0]
		rolesStr := parts[1]
		roles := strings.Split(rolesStr, ",")

		hostInfo := HostInfo{
			Hostname: hostname,
			Proposer: []int64{},
			Acceptor: []int64{},
			Learner:  []int64{},
		}

		for _, role := range roles {
			roleParts := strings.Split(role, ProposerRole)
			if len(roleParts) == 2 {
				num, err := strconv.ParseInt(roleParts[1], 10, 64)
				if err != nil {
					log.Fatalf("invalid role number in hostfile for %s: %v", role, err)
				}
				if num != 0 {
					hostInfo.Proposer = append(hostInfo.Proposer, num)
					// Add this proposer to the quorum map with an empty acceptor list (to be populated later)
					quorumMap[num] = []int64{}
				}
				continue
			}

			roleParts = strings.Split(role, AcceptorRole)
			if len(roleParts) == 2 {
				num, err := strconv.ParseInt(roleParts[1], 10, 64)
				if err != nil {
					log.Fatalf("invalid role number in hostfile for %s: %v", role, err)
				}
				if num != 0 {
					hostInfo.Acceptor = append(hostInfo.Acceptor, num)
				}
				continue
			}

			roleParts = strings.Split(role, LearnerRole)
			if len(roleParts) == 2 {
				num, err := strconv.ParseInt(roleParts[1], 10, 64)
				if err != nil {
					log.Fatalf("invalid role number in hostfile for %s: %v", role, err)
				}
				if num != 0 {
					hostInfo.Learner = append(hostInfo.Learner, num)
				}
				continue
			}

			log.Fatalf("unknown role in hostfile: %s", role)
		}

		hostRoles[lineID] = hostInfo
		lineID++
	}

	// Populate quorumMap with acceptors for each proposer
	for id, hostInfo := range hostRoles {
		for _, proposerID := range hostInfo.Proposer {
			for acceptorID, acceptorInfo := range hostRoles {
				if id != acceptorID && len(acceptorInfo.Acceptor) > 0 {
					quorumMap[proposerID] = append(quorumMap[proposerID], acceptorID)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("error reading hostfile: %v", err)
	}

	return hostRoles, quorumMap
}
