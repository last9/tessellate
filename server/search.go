package server

import (
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"strings"
)

type cloudResource struct {
	identifier []string
	aliCloud map[string]string
	gcp map[string]string
	aws map[string]string
}

func checkResource(resource string) (*cloudResource, error) {
	cr := cloudResource{}
	switch resource {
	case "cidr_range":
		cr.identifier = []string{"alicloud_vpc", "aws_vpc", "google_compute_subnetwork"}
		cr.aliCloud = map[string]string{
			"identifier": "alicloud_vpc",
			"key_path":"primary.attributes.cidr_block",
		}
		cr.aws = map[string]string{
			"identifier": "aws_vpc",
			"key_path":"primary.attributes.cidr_block",
		}
		cr.gcp = map[string]string{
			"identifier": "google_compute_subnetwork",
			"key_path":"primary.attributes.ip_cidr_range",
		}
	case "private_ip":
		cr.identifier = []string{"alicloud_instance", "aws_instance", "google_compute_instance"}
		cr.aliCloud = map[string]string{
			"identifier": "alicloud_instance",
			"key_path":"primary.attributes.private_ip",
		}
		cr.aws = map[string]string{
			"identifier": "aws_instance",
			"key_path":"primary.attributes.private_ip",
		}
		cr.gcp = map[string]string{
			"identifier": "google_compute_instance",
			"key_path":"primary.attributes.network_interface\\.0\\.network_ip",
		}
	case "public_ip":
		cr.identifier = []string{"alicloud_instance", "aws_instance", "google_compute_instance"}
		cr.aliCloud = map[string]string{
			"identifier": "alicloud_instance",
			"key_path":"primary.attributes.public_ip",
		}
		cr.aws = map[string]string{
			"identifier": "aws_instance",
			"key_path":"primary.attributes.public_ip",
		}
		cr.gcp = map[string]string{
			"identifier": "google_compute_instance",
			"key_path":"primary.attributes.network_interface\\.0\\.access_config\\.0\\.nat_ip",
		}
	default:
		return nil, errors.New("Invalid Resource passed. Currently tessellate supports either of " +
			"[cidr_range, private_ip, public_ip]")
	}
	return &cr, nil
}

func (cr *cloudResource) getResource(data []byte) []string {
	var resourceArray []string
	jsonParse := gjson.Get(string(data), "modules.0.resources")
	jsonParse.ForEach(func(key, val gjson.Result) bool {
		for _, j := range cr.identifier {
			if !strings.HasPrefix(key.String(), j) {
				continue
			}
			var v string
			switch j {
			case cr.aliCloud["identifier"]:
				v = gjson.Get(val.String(), cr.aliCloud["key_path"]).String()
			case cr.gcp["identifier"]:
				v = gjson.Get(val.String(), cr.gcp["key_path"]).String()
			case cr.aws["identifier"]:
				v = gjson.Get(val.String(), cr.aws["key_path"]).String()
			}
			if len(v) == 0 {
				continue
			}
			resourceArray = append(resourceArray, v)
		}
		return true
	})
	return resourceArray
}
