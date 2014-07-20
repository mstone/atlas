// Copyright (c) 2013, 2014 Akamai Technologies, Inc.

package sitelistcache

import (
	"log"
	"os"
	"path"
	"testing"
)

var testPath string
var chartsPath string

var siteListCache *SiteListCache

func init() {
	testPath := os.Getenv("ATLAS_TEST_PATH")

	if testPath == "" {
		testPath = "../"
	}

	chartsPath = path.Join(testPath, "test/charts")

	siteListCache = New(chartsPath)
}

func TestSiteListCacheMake(t *testing.T) {
	t.Parallel()

	built, err := siteListCache.Make()
	if err != nil {
		t.Fatalf("TestSiteListCacheMake() make failed: err: %q", err)
	}
	if !built {
		t.Fatalf("TestSiteListCacheMake() make did not build!")
	}

	built, err = siteListCache.Make()
	if err != nil {
		t.Fatalf("TestSiteListCacheMake() remake failed: err: %q", err)
	}

	if built {
		t.Fatalf("TestSiteListCacheMake() remake failed: something chnaged!", err)
	}

	for k, v := range siteListCache.Entries {
		log.Printf("TestSiteListCacheMake(): ent %q -> %v", k, v)
	}
}
