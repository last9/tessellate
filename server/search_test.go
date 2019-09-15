package server

import (
	"context"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tsocial/tessellate/storage/types"
	"io/ioutil"
	"testing"
)

func TestEmptySearchResource(t *testing.T) {
	t.Run("Send a resource to search, when no state file is in storage", func(t *testing.T) {
		req := &SearchRequest{ResourceName:"cidr_range"}
		v, err := server.SearchResource(context.Background(), req)
		assert.Nil(t, err, v)
	})
}

func TestSearchWithWrongResource(t *testing.T) {

	workspace := uuid.NewV4().String()
	layout := uuid.NewV4().String()

	key := fmt.Sprintf("%s/%s/%s", types.STATE, workspace, layout)
	valBytes, err := ioutil.ReadFile("./testdata/aws.tfstate")
	assert.Nil(t, err)

	err = store.SaveKey(key, valBytes)
	assert.Nil(t, err)

	req1 := &SearchRequest{ResourceName:"cidr_rang"}
	_, err1 := server.SearchResource(context.Background(), req1)
	assert.Contains(t, err1.Error(), "Invalid Resource passed. Currently tessellate supports either of [cidr_range, private_ip, public_ip]")
}

func TestSearchWithAWSStateFile(t *testing.T) {
	workspace := uuid.NewV4().String()
	layout := uuid.NewV4().String()

	key := fmt.Sprintf("%s/%s/%s", types.STATE, workspace, layout)
	valBytes, err := ioutil.ReadFile("./testdata/aws.tfstate")
	assert.Nil(t, err)

	err = store.SaveKey(key, valBytes)
	assert.Nil(t, err)

	req := &SearchRequest{ResourceName:"cidr_range"}
	v, err := server.SearchResource(context.Background(), req)
	assert.Equal(t, v.Resources[0].Resource[0], "10.28.98.0/24")
	assert.Equal(t, v.Resources[0].Workspace, workspace)
	assert.Equal(t, v.Resources[0].Layout, layout)

	req1 := &SearchRequest{ResourceName:"private_ip"}
	v1, _ := server.SearchResource(context.Background(), req1)
	assert.Equal(t, len(v1.Resources[0].Resource), 2)
	assert.ElementsMatch(t, v1.Resources[0].Resource, []string{"10.28.98.89", "10.28.98.37"})
	assert.Equal(t, v1.Resources[0].Workspace, workspace)
	assert.Equal(t, v1.Resources[0].Layout, layout)

	req2 := &SearchRequest{ResourceName:"public_ip"}
	v2, _ := server.SearchResource(context.Background(), req2)
	assert.ElementsMatch(t, v2.Resources[0].Resource, []string{"13.251.103.176", "54.169.227.200"})
	assert.Equal(t, v2.Resources[0].Workspace, workspace)
	assert.Equal(t, v2.Resources[0].Layout, layout)
}

func TestSearchWithAliCloudStateFile(t *testing.T) {
	workspace := uuid.NewV4().String()
	layout := uuid.NewV4().String()

	key := fmt.Sprintf("%s/%s/%s", types.STATE, workspace, layout)
	valBytes, err := ioutil.ReadFile("./testdata/alicloud.tfstate")
	assert.Nil(t, err)

	err = store.SaveKey(key, valBytes)
	assert.Nil(t, err)

	req := &SearchRequest{ResourceName:"cidr_range"}
	v, err := server.SearchResource(context.Background(), req)
	assert.Equal(t, v.Resources[0].Resource[0], "172.31.0.0/16")
	assert.Equal(t, v.Resources[0].Workspace, workspace)
	assert.Equal(t, v.Resources[0].Layout, layout)

	req1 := &SearchRequest{ResourceName:"private_ip"}
	v1, _ := server.SearchResource(context.Background(), req1)
	assert.Equal(t, len(v1.Resources[0].Resource), 2)
	assert.ElementsMatch(t, v1.Resources[0].Resource, []string{"172.31.8.117", "172.31.8.118"})
	assert.Equal(t, v1.Resources[0].Workspace, workspace)
	assert.Equal(t, v1.Resources[0].Layout, layout)

	req2 := &SearchRequest{ResourceName:"public_ip"}
	v2, _ := server.SearchResource(context.Background(), req2)
	assert.ElementsMatch(t, v2.Resources[0].Resource, []string{"148.229.221.103"})
	assert.Equal(t, v2.Resources[0].Workspace, workspace)
	assert.Equal(t, v2.Resources[0].Layout, layout)
}

func TestSearchWithGCPStateFile(t *testing.T) {
	workspace := uuid.NewV4().String()
	layout := uuid.NewV4().String()

	key := fmt.Sprintf("%s/%s/%s", types.STATE, workspace, layout)
	valBytes, err := ioutil.ReadFile("./testdata/gcp.tfstate")
	assert.Nil(t, err)

	err = store.SaveKey(key, valBytes)
	assert.Nil(t, err)

	req := &SearchRequest{ResourceName:"cidr_range"}
	v, err := server.SearchResource(context.Background(), req)
	assert.Equal(t, v.Resources[0].Resource, []string{"172.30.8.0/27","172.30.8.128/27"})
	assert.Equal(t, v.Resources[0].Workspace, workspace)
	assert.Equal(t, v.Resources[0].Layout, layout)

	req1 := &SearchRequest{ResourceName:"private_ip"}
	v1, _ := server.SearchResource(context.Background(), req1)
	assert.Equal(t, len(v1.Resources[0].Resource), 3)
	assert.ElementsMatch(t, v1.Resources[0].Resource, []string{"172.30.8.9", "172.30.8.157", "172.30.8.75"})
	assert.Equal(t, v1.Resources[0].Workspace, workspace)
	assert.Equal(t, v1.Resources[0].Layout, layout)

	req2 := &SearchRequest{ResourceName:"public_ip"}
	v2, _ := server.SearchResource(context.Background(), req2)
	assert.ElementsMatch(t, v2.Resources[0].Resource, []string{"34.245.149.7", "37.245.149.7"})
	assert.Equal(t, v2.Resources[0].Workspace, workspace)
	assert.Equal(t, v2.Resources[0].Layout, layout)
}

func TestSearchWithMultipleWorkSpaceLayouts(t *testing.T) {
	workspaceGCP := uuid.NewV4().String()
	layoutGCP := uuid.NewV4().String()

	keyGCP := fmt.Sprintf("%s/%s/%s", types.STATE, workspaceGCP, layoutGCP)
	valBytes, err := ioutil.ReadFile("./testdata/gcp.tfstate")
	assert.Nil(t, err)

	err = store.SaveKey(keyGCP, valBytes)
	assert.Nil(t, err)

	workspaceAWS := uuid.NewV4().String()
	layoutAWS := uuid.NewV4().String()

	keyAWS := fmt.Sprintf("%s/%s/%s", types.STATE, workspaceAWS, layoutAWS)
	valBytesAWS, err := ioutil.ReadFile("./testdata/aws.tfstate")
	assert.Nil(t, err)

	err = store.SaveKey(keyAWS, valBytesAWS)
	assert.Nil(t, err)

	workspaceAliCloud := uuid.NewV4().String()
	layoutAliCloud := uuid.NewV4().String()

	keyAliCloud := fmt.Sprintf("%s/%s/%s", types.STATE, workspaceAliCloud, layoutAliCloud)
	valBytesAliCloud, err := ioutil.ReadFile("./testdata/alicloud.tfstate")
	assert.Nil(t, err)

	err = store.SaveKey(keyAliCloud, valBytesAliCloud)
	assert.Nil(t, err)

	req := &SearchRequest{ResourceName:"cidr_range"}
	v, err := server.SearchResource(context.Background(), req)
	fmt.Println(v.Resources)
	assert.Equal(t, len(v.Resources), 3)
}
