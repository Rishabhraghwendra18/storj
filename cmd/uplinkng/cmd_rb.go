// Copyright (C) 2021 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"fmt"
	"strconv"

	"github.com/zeebo/clingy"
	"github.com/zeebo/errs"

	"storj.io/storj/cmd/uplinkng/ulext"
	"storj.io/storj/cmd/uplinkng/ulloc"
)

type cmdRb struct {
	ex ulext.External

	access string
	force  bool

	loc ulloc.Location
}

func newCmdRb(ex ulext.External) *cmdRb {
	return &cmdRb{ex: ex}
}

func (c *cmdRb) Setup(params clingy.Parameters) {
	c.access = params.Flag("access", "Access name or value to use", "").(string)
	c.force = params.Flag("force", "Deletes any objects in bucket first", false,
		clingy.Transform(strconv.ParseBool),
	).(bool)

	c.loc = params.Arg("name", "Bucket name (sj://BUCKET)",
		clingy.Transform(ulloc.Parse),
	).(ulloc.Location)
}

func (c *cmdRb) Execute(ctx clingy.Context) error {
	project, err := c.ex.OpenProject(ctx, c.access)
	if err != nil {
		return err
	}
	defer func() { _ = project.Close() }()

	bucket, key, ok := c.loc.RemoteParts()
	if !ok {
		return errs.New("location must be remote")
	}
	if key != "" {
		return errs.New("key must not be specified: %q", key)
	}

	if c.force {
		_, err = project.DeleteBucketWithObjects(ctx, bucket)
	} else {
		_, err = project.DeleteBucket(ctx, bucket)
	}
	if err != nil {
		return err
	}

	fmt.Fprintf(ctx.Stdout(), "Bucket %q has been deleted.\n", bucket)
	return nil
}
