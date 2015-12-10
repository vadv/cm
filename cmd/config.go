package main

import (
	"encoding/json"
	"fmt"

	"cm/log"
	"cm/receiver"
	"cm/sender"
	"cm/storage"
	"cm/supervisor"
)

func readConfigFile(filename string, log *log.Logger) (*BusConfig, error) {

	busConfig := newBusConfig(log)
	busConfig.setCommonStorage(storage.NewQueueFifo())

	confFile, err := parseFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Error read config file: %s, %s", filename, err.Error())
	}

	for pluginName, pluginSection := range confFile {

		pluginType, ok := pluginSection["type"]
		if !ok {
			return nil, fmt.Errorf("Unknown 'type' for: %s", pluginName)
		}

		pluginConfig, err := json.Marshal(pluginSection)
		if err != nil {
			return nil, fmt.Errorf("Json marshal %v: %s", pluginSection, err.Error())
		}

		busConfig.setSettings(pluginName, pluginConfig)

		// пока тут все через опу.
		switch pluginType {
		case "opentsdb":
			if opentsdb, err := sender.NewOpenTsdbSender(pluginName, busConfig); err != nil {
				return nil, fmt.Errorf("create opentsdb sender %#v: %s", pluginName, err.Error())
			} else {
				busConfig.addNewSender(pluginName, opentsdb)
			}
		case "graphite":
			if graphite, err := sender.NewGrahiteSender(pluginName, busConfig); err != nil {
				return nil, fmt.Errorf("create graphite sender %#v: %s", pluginName, err.Error())
			} else {
				busConfig.addNewSender(pluginName, graphite)
			}
		case "zabbix":
			if zabbix, err := sender.NewZabbixSender(pluginName, busConfig); err != nil {
				return nil, fmt.Errorf("create zabbix sender %#v: %s", pluginName, err.Error())
			} else {
				busConfig.addNewSender(pluginName, zabbix)
			}
		case "http_reciever":
			if http_reciever, err := receiver.NewHTTPReceiver(pluginName, busConfig); err != nil {
				return nil, fmt.Errorf("create http reciever sender %#v: %s", pluginName, err.Error())
			} else {
				busConfig.addNewReciever(pluginName, http_reciever)
			}
		case "supervisor":
			if task, err := supervisor.NewTask(pluginName, busConfig); err != nil {
				return nil, fmt.Errorf("create supervisor %#v: %s", pluginName, err.Error())
			} else {
				busConfig.addSupervisorTasks(pluginName, task)
			}
		default:
			return nil, fmt.Errorf("unknown type: %#v for name: %#v", pluginType, pluginName)
		}
	}

	return busConfig, nil
}
