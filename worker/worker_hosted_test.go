// +build exectest,buildtest

package worker_test

import (
	"testing"

	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/server/testserver"
	"src.sourcegraph.com/sourcegraph/util/testutil"
)

func TestBuildRepo_serverside_hosted_lg(t *testing.T) {
	if testserver.Store == "pgsql" {
		t.Skip()
	}

	t.Parallel()

	a, ctx := testserver.NewUnstartedServer()
	a.Config.ServeFlags = append(a.Config.ServeFlags,
		&authutil.Flags{Source: "none", AllowAnonymousReaders: true},
	)
	if err := a.Start(); err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	commitID, close, err := testutil.CreateAndPushRepo(t, ctx, "r/r")
	if err != nil {
		t.Fatal(err)
	}
	defer close()

	// Pushing triggers a build; wait for it to finish.
	build, err := testutil.WaitForBuild(t, ctx, sourcegraph.BuildSpec{Repo: sourcegraph.RepoSpec{URI: "r/r"}, CommitID: commitID, Attempt: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !build.Success {
		t.Fatalf("build %s failed", build.Spec().IDString())
	}

	checkImport(t, ctx, a.Client, "r/r", commitID)
}

func TestBuildRepo_push_hosted_lg(t *testing.T) {
	if testserver.Store == "pgsql" {
		t.Skip()
	}

	t.Parallel()

	a, ctx := testserver.NewUnstartedServer()
	a.Config.ServeFlags = append(a.Config.ServeFlags,
		&authutil.Flags{Source: "none", AllowAnonymousReaders: true},
	)
	if err := a.Start(); err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	_, close, err := testutil.CreateAndPushRepo(t, ctx, "r/rr")
	if err != nil {
		t.Fatal(err)
	}
	defer close()

	repo, err := a.Client.Repos.Get(ctx, &sourcegraph.RepoSpec{URI: "r/rr"})
	if err != nil {
		t.Fatal(err)
	}

	// Clone and build the repo locally.
	if err := cloneAndLocallyBuildRepo(t, a, repo, ""); err != nil {
		t.Fatal(err)
	}

	checkImport(t, ctx, a.Client, "r/rr", "")
}
