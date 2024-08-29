package main

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/karlsen-network/karlsend/v2/infrastructure/network/netadapter/server/grpcserver/protowire"
)

var commandTypes = []reflect.Type{
	reflect.TypeOf(protowire.KarlsendMessage_AddPeerRequest{}),
	reflect.TypeOf(protowire.KarlsendMessage_GetConnectedPeerInfoRequest{}),
	reflect.TypeOf(protowire.KarlsendMessage_GetPeerAddressesRequest{}),
	reflect.TypeOf(protowire.KarlsendMessage_GetCurrentNetworkRequest{}),
	reflect.TypeOf(protowire.KarlsendMessage_GetInfoRequest{}),

	reflect.TypeOf(protowire.KarlsendMessage_GetBlockRequest{}),
	reflect.TypeOf(protowire.KarlsendMessage_GetBlocksRequest{}),
	reflect.TypeOf(protowire.KarlsendMessage_GetHeadersRequest{}),
	reflect.TypeOf(protowire.KarlsendMessage_GetBlockCountRequest{}),
	reflect.TypeOf(protowire.KarlsendMessage_GetBlockDagInfoRequest{}),
	reflect.TypeOf(protowire.KarlsendMessage_GetSelectedTipHashRequest{}),
	reflect.TypeOf(protowire.KarlsendMessage_GetVirtualSelectedParentBlueScoreRequest{}),
	reflect.TypeOf(protowire.KarlsendMessage_GetVirtualSelectedParentChainFromBlockRequest{}),
	reflect.TypeOf(protowire.KarlsendMessage_ResolveFinalityConflictRequest{}),
	reflect.TypeOf(protowire.KarlsendMessage_EstimateNetworkHashesPerSecondRequest{}),

	reflect.TypeOf(protowire.KarlsendMessage_GetBlockTemplateRequest{}),
	reflect.TypeOf(protowire.KarlsendMessage_SubmitBlockRequest{}),

	reflect.TypeOf(protowire.KarlsendMessage_GetMempoolEntryRequest{}),
	reflect.TypeOf(protowire.KarlsendMessage_GetMempoolEntriesRequest{}),
	reflect.TypeOf(protowire.KarlsendMessage_GetMempoolEntriesByAddressesRequest{}),

	reflect.TypeOf(protowire.KarlsendMessage_SubmitTransactionRequest{}),

	reflect.TypeOf(protowire.KarlsendMessage_GetUtxosByAddressesRequest{}),
	reflect.TypeOf(protowire.KarlsendMessage_GetBalanceByAddressRequest{}),
	reflect.TypeOf(protowire.KarlsendMessage_GetCoinSupplyRequest{}),

	reflect.TypeOf(protowire.KarlsendMessage_BanRequest{}),
	reflect.TypeOf(protowire.KarlsendMessage_UnbanRequest{}),
}

type commandDescription struct {
	name       string
	parameters []*parameterDescription
	typeof     reflect.Type
}

type parameterDescription struct {
	name   string
	typeof reflect.Type
}

func commandDescriptions() []*commandDescription {
	commandDescriptions := make([]*commandDescription, len(commandTypes))

	for i, commandTypeWrapped := range commandTypes {
		commandType := unwrapCommandType(commandTypeWrapped)

		name := strings.TrimSuffix(commandType.Name(), "RequestMessage")
		numFields := commandType.NumField()

		var parameters []*parameterDescription
		for i := 0; i < numFields; i++ {
			field := commandType.Field(i)

			if !isFieldExported(field) {
				continue
			}

			parameters = append(parameters, &parameterDescription{
				name:   field.Name,
				typeof: field.Type,
			})
		}
		commandDescriptions[i] = &commandDescription{
			name:       name,
			parameters: parameters,
			typeof:     commandTypeWrapped,
		}
	}

	return commandDescriptions
}

func (cd *commandDescription) help() string {
	sb := &strings.Builder{}
	sb.WriteString(cd.name)
	for _, parameter := range cd.parameters {
		_, _ = fmt.Fprintf(sb, " [%s]", parameter.name)
	}
	return sb.String()
}
