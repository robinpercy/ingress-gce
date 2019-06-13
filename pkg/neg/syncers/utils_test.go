/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package syncers

import (
	"reflect"
	"strconv"
	"testing"

	"google.golang.org/api/compute/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	negtypes "k8s.io/ingress-gce/pkg/neg/types"
)

func TestEncodeDecodeEndpoint(t *testing.T) {
	ip := "10.0.0.10"
	instance := "somehost"
	port := "8080"

	retIp, retInstance, retPort := decodeEndpoint(encodeEndpoint(ip, instance, port))

	if ip != retIp || instance != retInstance || retPort != port {
		t.Fatalf("Encode and decode endpoint failed. Expect %q, %q, %q but got %q, %q, %q.", ip, instance, port, retIp, retInstance, retPort)
	}
}

func TestCalculateDifference(t *testing.T) {
	testCases := []struct {
		targetSet  map[string]sets.String
		currentSet map[string]sets.String
		addSet     map[string]sets.String
		removeSet  map[string]sets.String
	}{
		// unchanged
		{
			targetSet: map[string]sets.String{
				negtypes.TestZone1: sets.NewString("a", "b", "c"),
			},
			currentSet: map[string]sets.String{
				negtypes.TestZone1: sets.NewString("a", "b", "c"),
			},
			addSet:    map[string]sets.String{},
			removeSet: map[string]sets.String{},
		},
		// unchanged
		{
			targetSet:  map[string]sets.String{},
			currentSet: map[string]sets.String{},
			addSet:     map[string]sets.String{},
			removeSet:  map[string]sets.String{},
		},
		// add in one zone
		{
			targetSet: map[string]sets.String{
				negtypes.TestZone1: sets.NewString("a", "b", "c"),
			},
			currentSet: map[string]sets.String{},
			addSet: map[string]sets.String{
				negtypes.TestZone1: sets.NewString("a", "b", "c"),
			},
			removeSet: map[string]sets.String{},
		},
		// add in 2 zones
		{
			targetSet: map[string]sets.String{
				negtypes.TestZone1: sets.NewString("a", "b", "c"),
				negtypes.TestZone2: sets.NewString("e", "f", "g"),
			},
			currentSet: map[string]sets.String{},
			addSet: map[string]sets.String{
				negtypes.TestZone1: sets.NewString("a", "b", "c"),
				negtypes.TestZone2: sets.NewString("e", "f", "g"),
			},
			removeSet: map[string]sets.String{},
		},
		// remove in one zone
		{
			targetSet: map[string]sets.String{},
			currentSet: map[string]sets.String{
				negtypes.TestZone1: sets.NewString("a", "b", "c"),
			},
			addSet: map[string]sets.String{},
			removeSet: map[string]sets.String{
				negtypes.TestZone1: sets.NewString("a", "b", "c"),
			},
		},
		// remove in 2 zones
		{
			targetSet: map[string]sets.String{},
			currentSet: map[string]sets.String{
				negtypes.TestZone1: sets.NewString("a", "b", "c"),
				negtypes.TestZone2: sets.NewString("e", "f", "g"),
			},
			addSet: map[string]sets.String{},
			removeSet: map[string]sets.String{
				negtypes.TestZone1: sets.NewString("a", "b", "c"),
				negtypes.TestZone2: sets.NewString("e", "f", "g"),
			},
		},
		// add and delete in one zone
		{
			targetSet: map[string]sets.String{
				negtypes.TestZone1: sets.NewString("a", "b", "c"),
			},
			currentSet: map[string]sets.String{
				negtypes.TestZone1: sets.NewString("b", "c", "d"),
			},
			addSet: map[string]sets.String{
				negtypes.TestZone1: sets.NewString("a"),
			},
			removeSet: map[string]sets.String{
				negtypes.TestZone1: sets.NewString("d"),
			},
		},
		// add and delete in 2 zones
		{
			targetSet: map[string]sets.String{
				negtypes.TestZone1: sets.NewString("a", "b", "c"),
				negtypes.TestZone2: sets.NewString("a", "b", "c"),
			},
			currentSet: map[string]sets.String{
				negtypes.TestZone1: sets.NewString("b", "c", "d"),
				negtypes.TestZone2: sets.NewString("b", "c", "d"),
			},
			addSet: map[string]sets.String{
				negtypes.TestZone1: sets.NewString("a"),
				negtypes.TestZone2: sets.NewString("a"),
			},
			removeSet: map[string]sets.String{
				negtypes.TestZone1: sets.NewString("d"),
				negtypes.TestZone2: sets.NewString("d"),
			},
		},
	}

	for _, tc := range testCases {
		addSet, removeSet := calculateDifference(tc.targetSet, tc.currentSet)

		if !reflect.DeepEqual(addSet, tc.addSet) {
			t.Errorf("Failed to calculate difference for add, expecting %v, but got %v", tc.addSet, addSet)
		}

		if !reflect.DeepEqual(removeSet, tc.removeSet) {
			t.Errorf("Failed to calculate difference for remove, expecting %v, but got %v", tc.removeSet, removeSet)
		}
	}
}

func TestNetworkEndpointCalculateDifference(t *testing.T) {
	testCases := []struct {
		targetSet  map[string]negtypes.NetworkEndpointSet
		currentSet map[string]negtypes.NetworkEndpointSet
		addSet     map[string]negtypes.NetworkEndpointSet
		removeSet  map[string]negtypes.NetworkEndpointSet
	}{
		// unchanged
		{
			targetSet: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("a"), genNetworkEndpoint("b"), genNetworkEndpoint("c")),
			},
			currentSet: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("a"), genNetworkEndpoint("b"), genNetworkEndpoint("c")),
			},
			addSet:    map[string]negtypes.NetworkEndpointSet{},
			removeSet: map[string]negtypes.NetworkEndpointSet{},
		},
		// unchanged
		{
			targetSet:  map[string]negtypes.NetworkEndpointSet{},
			currentSet: map[string]negtypes.NetworkEndpointSet{},
			addSet:     map[string]negtypes.NetworkEndpointSet{},
			removeSet:  map[string]negtypes.NetworkEndpointSet{},
		},
		// add in one zone
		{
			targetSet: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("a"), genNetworkEndpoint("b"), genNetworkEndpoint("c")),
			},
			currentSet: map[string]negtypes.NetworkEndpointSet{},
			addSet: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("a"), genNetworkEndpoint("b"), genNetworkEndpoint("c")),
			},
			removeSet: map[string]negtypes.NetworkEndpointSet{},
		},
		// add in 2 zones
		{
			targetSet: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("a"), genNetworkEndpoint("b"), genNetworkEndpoint("c")),
				negtypes.TestZone2: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("e"), genNetworkEndpoint("f"), genNetworkEndpoint("g")),
			},
			currentSet: map[string]negtypes.NetworkEndpointSet{},
			addSet: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("a"), genNetworkEndpoint("b"), genNetworkEndpoint("c")),
				negtypes.TestZone2: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("e"), genNetworkEndpoint("f"), genNetworkEndpoint("g")),
			},
			removeSet: map[string]negtypes.NetworkEndpointSet{},
		},
		// remove in one zone
		{
			targetSet: map[string]negtypes.NetworkEndpointSet{},
			currentSet: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("a"), genNetworkEndpoint("b"), genNetworkEndpoint("c")),
			},
			addSet: map[string]negtypes.NetworkEndpointSet{},
			removeSet: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("a"), genNetworkEndpoint("b"), genNetworkEndpoint("c")),
			},
		},
		// remove in 2 zones
		{
			targetSet: map[string]negtypes.NetworkEndpointSet{},
			currentSet: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("a"), genNetworkEndpoint("b"), genNetworkEndpoint("c")),
				negtypes.TestZone2: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("e"), genNetworkEndpoint("f"), genNetworkEndpoint("g")),
			},
			addSet: map[string]negtypes.NetworkEndpointSet{},
			removeSet: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("a"), genNetworkEndpoint("b"), genNetworkEndpoint("c")),
				negtypes.TestZone2: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("e"), genNetworkEndpoint("f"), genNetworkEndpoint("g")),
			},
		},
		// add and delete in one zone
		{
			targetSet: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("a"), genNetworkEndpoint("b"), genNetworkEndpoint("c")),
			},
			currentSet: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("b"), genNetworkEndpoint("c"), genNetworkEndpoint("d")),
			},
			addSet: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("a")),
			},
			removeSet: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("d")),
			},
		},
		// add and delete in 2 zones
		{
			targetSet: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("a"), genNetworkEndpoint("b"), genNetworkEndpoint("c")),
				negtypes.TestZone2: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("a"), genNetworkEndpoint("b"), genNetworkEndpoint("c")),
			},
			currentSet: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("b"), genNetworkEndpoint("c"), genNetworkEndpoint("d")),
				negtypes.TestZone2: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("b"), genNetworkEndpoint("c"), genNetworkEndpoint("d")),
			},
			addSet: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("a")),
				negtypes.TestZone2: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("a")),
			},
			removeSet: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("d")),
				negtypes.TestZone2: negtypes.NewNetworkEndpointSet(genNetworkEndpoint("d")),
			},
		},
	}

	for _, tc := range testCases {
		addSet, removeSet := calculateNetworkEndpointDifference(tc.targetSet, tc.currentSet)

		if !reflect.DeepEqual(addSet, tc.addSet) {
			t.Errorf("Failed to calculate difference for add, expecting %v, but got %v", tc.addSet, addSet)
		}

		if !reflect.DeepEqual(removeSet, tc.removeSet) {
			t.Errorf("Failed to calculate difference for remove, expecting %v, but got %v", tc.removeSet, removeSet)
		}
	}
}

// TODO(freehan): add test cases with Endpoints with NotReady addresses
func TestToZoneNetworkEndpointMapUtil(t *testing.T) {
	zoneGetter := negtypes.NewFakeZoneGetter()
	testCases := []struct {
		targetPort string
		expect     map[string]negtypes.NetworkEndpointSet
	}{
		// Non exist
		{
			targetPort: "8888",
			expect:     map[string]negtypes.NetworkEndpointSet{},
		},
		{
			targetPort: "80",
			expect: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(networkEndpointFromEncodedEndpoint("10.100.1.1||instance1||80"), networkEndpointFromEncodedEndpoint("10.100.1.2||instance1||80"), networkEndpointFromEncodedEndpoint("10.100.2.1||instance2||80")),
				negtypes.TestZone2: negtypes.NewNetworkEndpointSet(networkEndpointFromEncodedEndpoint("10.100.3.1||instance3||80")),
			},
		},
		{
			targetPort: testNamedPort,
			expect: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(networkEndpointFromEncodedEndpoint("10.100.2.2||instance2||81")),
				negtypes.TestZone2: negtypes.NewNetworkEndpointSet(networkEndpointFromEncodedEndpoint("10.100.4.1||instance4||81"), networkEndpointFromEncodedEndpoint("10.100.3.2||instance3||8081"), networkEndpointFromEncodedEndpoint("10.100.4.2||instance4||8081")),
			},
		},
	}

	for _, tc := range testCases {
		res, _, _ := toZoneNetworkEndpointMap(getDefaultEndpoint(), zoneGetter, tc.targetPort, nil)

		if !reflect.DeepEqual(res, tc.expect) {
			t.Errorf("Expect %v, but got %v.", tc.expect, res)
		}
	}
}

func TestRetrieveExistingZoneNetworkEndpointMap(t *testing.T) {
	zoneGetter := negtypes.NewFakeZoneGetter()
	negCloud := negtypes.NewFakeNetworkEndpointGroupCloud("test-subnetwork", "test-newtork")
	negName := "test-neg-name"
	irrelevantNegName := "irrelevant"
	testIP1 := "1.2.3.4"
	testIP2 := "1.2.3.5"
	testIP3 := "1.2.3.6"
	testIP4 := "1.2.3.7"
	testPort := int64(80)

	testCases := []struct {
		desc      string
		mutate    func(cloud negtypes.NetworkEndpointGroupCloud)
		expect    map[string]negtypes.NetworkEndpointSet
		expectErr bool
	}{
		{
			desc:      "neg not exists",
			mutate:    func(cloud negtypes.NetworkEndpointGroupCloud) {},
			expectErr: true,
		},
		{
			desc: "neg only exists in one of the zone",
			mutate: func(cloud negtypes.NetworkEndpointGroupCloud) {
				cloud.CreateNetworkEndpointGroup(&compute.NetworkEndpointGroup{Name: testNegName}, negtypes.TestZone1)
			},
			expectErr: true,
		},
		{
			desc: "neg only exists in one of the zone plus irrelevant negs",
			mutate: func(cloud negtypes.NetworkEndpointGroupCloud) {
				cloud.CreateNetworkEndpointGroup(&compute.NetworkEndpointGroup{Name: irrelevantNegName}, negtypes.TestZone2)
			},
			expectErr: true,
		},
		{
			desc: "empty negs exists in both zones",
			mutate: func(cloud negtypes.NetworkEndpointGroupCloud) {
				cloud.CreateNetworkEndpointGroup(&compute.NetworkEndpointGroup{Name: testNegName}, negtypes.TestZone2)
			},
			expect: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(),
				negtypes.TestZone2: negtypes.NewNetworkEndpointSet(),
			},
			expectErr: false,
		},
		{
			desc: "one empty and one non-empty negs",
			mutate: func(cloud negtypes.NetworkEndpointGroupCloud) {
				cloud.AttachNetworkEndpoints(testNegName, negtypes.TestZone1, []*compute.NetworkEndpoint{
					{
						Instance:  negtypes.TestInstance1,
						IpAddress: testIP1,
						Port:      testPort,
					},
				})
			},
			expect: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(negtypes.NetworkEndpoint{IP: testIP1, Node: negtypes.TestInstance1, Port: strconv.Itoa(int(testPort))}),
				negtypes.TestZone2: negtypes.NewNetworkEndpointSet(),
			},
			expectErr: false,
		},
		{
			desc: "one neg with multiple endpoints",
			mutate: func(cloud negtypes.NetworkEndpointGroupCloud) {
				cloud.AttachNetworkEndpoints(testNegName, negtypes.TestZone1, []*compute.NetworkEndpoint{
					{
						Instance:  negtypes.TestInstance2,
						IpAddress: testIP2,
						Port:      testPort,
					},
				})
			},
			expect: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(
					negtypes.NetworkEndpoint{IP: testIP1, Node: negtypes.TestInstance1, Port: strconv.Itoa(int(testPort))},
					negtypes.NetworkEndpoint{IP: testIP2, Node: negtypes.TestInstance2, Port: strconv.Itoa(int(testPort))},
				),
				negtypes.TestZone2: negtypes.NewNetworkEndpointSet(),
			},
			expectErr: false,
		},
		{
			desc: "both negs with multiple endpoints",
			mutate: func(cloud negtypes.NetworkEndpointGroupCloud) {
				cloud.AttachNetworkEndpoints(testNegName, negtypes.TestZone2, []*compute.NetworkEndpoint{
					{
						Instance:  negtypes.TestInstance3,
						IpAddress: testIP3,
						Port:      testPort,
					},
					{
						Instance:  negtypes.TestInstance4,
						IpAddress: testIP4,
						Port:      testPort,
					},
				})
			},
			expect: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(
					negtypes.NetworkEndpoint{IP: testIP1, Node: negtypes.TestInstance1, Port: strconv.Itoa(int(testPort))},
					negtypes.NetworkEndpoint{IP: testIP2, Node: negtypes.TestInstance2, Port: strconv.Itoa(int(testPort))},
				),
				negtypes.TestZone2: negtypes.NewNetworkEndpointSet(
					negtypes.NetworkEndpoint{IP: testIP3, Node: negtypes.TestInstance3, Port: strconv.Itoa(int(testPort))},
					negtypes.NetworkEndpoint{IP: testIP4, Node: negtypes.TestInstance4, Port: strconv.Itoa(int(testPort))},
				),
			},
			expectErr: false,
		},
		{
			desc: "irrelevant neg",
			mutate: func(cloud negtypes.NetworkEndpointGroupCloud) {
				cloud.AttachNetworkEndpoints(irrelevantNegName, negtypes.TestZone2, []*compute.NetworkEndpoint{
					{
						Instance:  negtypes.TestInstance3,
						IpAddress: testIP4,
						Port:      testPort,
					},
				})
			},
			expect: map[string]negtypes.NetworkEndpointSet{
				negtypes.TestZone1: negtypes.NewNetworkEndpointSet(
					negtypes.NetworkEndpoint{IP: testIP1, Node: negtypes.TestInstance1, Port: strconv.Itoa(int(testPort))},
					negtypes.NetworkEndpoint{IP: testIP2, Node: negtypes.TestInstance2, Port: strconv.Itoa(int(testPort))},
				),
				negtypes.TestZone2: negtypes.NewNetworkEndpointSet(
					negtypes.NetworkEndpoint{IP: testIP3, Node: negtypes.TestInstance3, Port: strconv.Itoa(int(testPort))},
					negtypes.NetworkEndpoint{IP: testIP4, Node: negtypes.TestInstance4, Port: strconv.Itoa(int(testPort))},
				),
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		tc.mutate(negCloud)
		out, err := retrieveExistingZoneNetworkEndpointMap(negName, zoneGetter, negCloud)

		if tc.expectErr {
			if err == nil {
				t.Errorf("For test case %q, expecting error, but got nil", tc.desc)
			}
		} else {
			if err != nil {
				t.Errorf("For test case %q, expect err = nil, but got %v", tc.desc, err)
			}
		}

		if !tc.expectErr {
			if !reflect.DeepEqual(out, tc.expect) {
				t.Errorf("For test case %q, expect output = %+v, but got %+v", tc.desc, tc.expect, out)
			}
		}
	}
}

func TestMakeEndpointBatch(t *testing.T) {
	testCases := []struct {
		desc        string
		endpointNum int
		leftOverNum int
	}{
		{
			"input with zero endpoints",
			0,
			0,
		},
		{
			"input with 1 endpoints",
			1,
			0,
		},
		{
			"input with 500 endpoints",
			500,
			0,
		},
		{
			"input with 501 endpoints",
			501,
			1,
		},
		{
			"input with 1000 endpoints",
			1000,
			500,
		},
	}

	for _, tc := range testCases {
		endpointSet, endpointMap := genTestEndpoints(tc.endpointNum)
		out, err := makeEndpointBatch(endpointSet)

		if err != nil {
			t.Errorf("Expect err = nil, but got %v", err)
		}

		if endpointSet.Len() != tc.leftOverNum {
			t.Errorf("Expect endpoint set has %d endpoints left, but got %d", tc.leftOverNum, endpointSet.Len())
		}

		expectOutputEndpoints := tc.endpointNum
		if tc.endpointNum > MAX_NETWORK_ENDPOINTS_PER_BATCH {
			expectOutputEndpoints = MAX_NETWORK_ENDPOINTS_PER_BATCH
		}

		if expectOutputEndpoints != len(out) {
			t.Errorf("Expect %d endpoint(s) in output, but got %d", expectOutputEndpoints, len(out))
		}

		for key, endpoint := range out {
			if endpointSet.Has(key) {
				t.Errorf("Expect %q endpoint to exist in output endpoint map, but not", key)
			}
			expectEndpoint, ok := endpointMap[key]
			if !ok {
				t.Errorf("Expect %q endpoint to exist in expected endpoint map, but not", key)
			} else {
				if !reflect.DeepEqual(expectEndpoint, endpoint) {
					t.Errorf("Expect endpoint object %+v, but got %+v", expectEndpoint, endpoint)
				}
			}
		}
	}
}

func genTestEndpoints(num int) (negtypes.NetworkEndpointSet, map[negtypes.NetworkEndpoint]*compute.NetworkEndpoint) {
	endpointSet := negtypes.NewNetworkEndpointSet()
	endpointMap := map[negtypes.NetworkEndpoint]*compute.NetworkEndpoint{}
	ip := "1.2.3.4"
	instance := "instance"
	for port := 0; port < num; port++ {
		key := negtypes.NetworkEndpoint{IP: ip, Node: instance, Port: strconv.Itoa(port)}
		endpointSet.Insert(key)
		endpointMap[key] = &compute.NetworkEndpoint{
			IpAddress: ip,
			Instance:  instance,
			Port:      int64(port),
		}
	}
	return endpointSet, endpointMap
}

func networkEndpointFromEncodedEndpoint(encodedEndpoint string) negtypes.NetworkEndpoint {
	ip, node, port := decodeEndpoint(encodedEndpoint)
	return negtypes.NetworkEndpoint{IP: ip, Node: node, Port: port}
}
