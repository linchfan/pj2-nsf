package nsfContext

import (
	"encoding/json"
	"fmt"
	"nsf/internal/db"
	"nsf/logger"
)

// Record the service's information.
type serviceInfo struct {
	NFName  string `json:"nfName"`
	SBIAddr string `json:"sbiAddr"`
	SBIPort int    `json:"sbiPort"`
}

// TODO: You may need to add some fields
type serviceContextIE struct {
	db.DbIf

	*logger.NsfLogger
}

// TODO: You may need to add some fields
type serviceContext struct {
	db.DbIf

	*logger.NsfLogger
}

// TODO: Finish the initialization of the serviceContext instance
func newServiceContext(serviceContextIE *serviceContextIE) *serviceContext {

	return &serviceContext{
		DbIf: serviceContextIE.DbIf,

		NsfLogger: serviceContextIE.NsfLogger,
	}
}

// TODO: Finish the release of the serviceContext instance
func (s *serviceContext) release() {
	s.NsfLogger.SvcLog.Infoln("Release serviceContext...")

	s.NsfLogger.SvcLog.Infoln("serviceContext released")
}

// TODO: Finish the implementation of the registerServices method with an error return type
func (s *serviceContext) registerServices(nfName, sbiAddr string, sbiPort int, services []string) error {
	s.NsfLogger.SvcLog.Tracef("Registering %s services", nfName)

	info := serviceInfo{
		NFName:  nfName,
		SBIAddr: sbiAddr,
		SBIPort: sbiPort,
	}

	// Marshal the service information into JSON.
	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal service info: %w", err)
	}

	// Map each service name to the serialized NF information.
	records := make(map[string]string, len(services))

	for _, service := range services {
		records[service] = string(data)
	}

	// Persist all registered services to the database.
	if err := s.DbIf.SaveAll(records); err != nil {
		return fmt.Errorf("failed to register services for %s: %w", nfName, err)
	}

	return nil
}

// TODO: Finish the implementation of the deregisterServices method with an error return type
func (s *serviceContext) deregisterServices(nfName string, services []string) error {
	s.NsfLogger.SvcLog.Tracef("De-registering %s services", nfName)

	for _, service := range services {
		// Check whether the target service exists.
		exists, err := s.DbIf.Exist(service)
		if err != nil {
			return fmt.Errorf("failed to check service %s: %w", service, err)
		}
		if !exists {
			return fmt.Errorf("service %s not found", service)
		}

		// Load the registered service information from the database.
		data, err := s.DbIf.Load(service)
		if err != nil {
			return fmt.Errorf("failed to load service %s: %w", service, err)
		}

		// Unmarshal the data string into serviceInfo struct.
		var info serviceInfo
		if err := json.Unmarshal([]byte(data), &info); err != nil {
			return fmt.Errorf("failed to decode service %s: %w", service, err)
		}

		// Verify that the service belongs to the requesting NF.
		if info.NFName != nfName {
			return fmt.Errorf("service %s does not belong to NF %s", service, nfName)
		}

		// Remove the service from the database.
		if err := s.DbIf.Remove(service); err != nil {
			return fmt.Errorf("failed to remove service %s: %w", service, err)
		}
	}

	return nil
}

// TODO: Finish the implementation of the discoverService method with these return values:
// (string, string, int, error) representing the NF name, SBI address, SBI port, and an error if any
func (s *serviceContext) discoverService(service string) (string, string, int, error) {
	s.NsfLogger.SvcLog.Tracef("Discovering service %s", service)

	// Check whether the target service exists.
	exists, err := s.DbIf.Exist(service)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to check service %s: %w", service, err)
	}
	if !exists {
		return "", "", 0, fmt.Errorf("service %s not found", service)
	}

	// Load the registered service information from the database.
	data, err := s.DbIf.Load(service)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to load service %s: %w", service, err)
	}

	// Unmarshal the data string into serviceInfo struct.
	var info serviceInfo
	if err := json.Unmarshal([]byte(data), &info); err != nil {
		return "", "", 0, fmt.Errorf("failed to decode service %s: %w", service, err)
	}

	// Return the NF name and SBI endpoint information.
	return info.NFName, info.SBIAddr, info.SBIPort, nil
}
