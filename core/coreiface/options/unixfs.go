package options

import (
	"errors"
	"fmt"

	dag "github.com/ipfs/boxo/ipld/merkledag"
	cid "github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
)

type Layout int

const (
	BalancedLayout Layout = iota
	TrickleLayout
)

type UnixfsAddSettings struct {
	CidVersion int
	MhType     uint64

	Inline       bool
	InlineLimit  int
	RawLeaves    bool
	RawLeavesSet bool

	Chunker string
	Layout  Layout

	Pin      bool
	OnlyHash bool
	FsCache  bool
	NoCopy   bool

	Events   chan<- interface{}
	Silent   bool
	Progress bool

	ToFiles string
}

type UnixfsLsSettings struct {
	ResolveChildren   bool
	UseCumulativeSize bool
}

type UnixfsMkdirSettings struct {
	Parents    bool
	CidVersion int
	MhType     uint64
}

type UnixfsRmSettings struct {
	Recursive bool
	Force     bool
}

type UnixfsCpSettings struct {
	Parents bool
}

type UnixfsReadSettings struct {
	Offset int64
	Count  int64
}

type UnixfsStatSettings struct {
	Format    string
	Hash      bool
	Size      bool
	WithLocal bool
}

type UnixfsWriteSettings struct {
	Offset    int64
	Create    bool
	Parents   bool
	Truncate  bool
	Count     int64
	RawLeaves bool

	CidVersion int
	MhType     uint64
}

type (
	UnixfsAddOption   func(*UnixfsAddSettings) error
	UnixfsLsOption    func(*UnixfsLsSettings) error
	UnixfsMkdirOption func(*UnixfsMkdirSettings) error
	UnixfsRmOption    func(*UnixfsRmSettings) error
	UnixfsCpOption    func(*UnixfsCpSettings) error
	UnixfsReadOption  func(*UnixfsReadSettings) error
	UnixfsStatOption  func(*UnixfsStatSettings) error
	UnixfsWriteOption func(*UnixfsWriteSettings) error
)

func UnixfsAddOptions(opts ...UnixfsAddOption) (*UnixfsAddSettings, cid.Prefix, error) {
	options := &UnixfsAddSettings{
		CidVersion: -1,
		MhType:     mh.SHA2_256,

		Inline:       false,
		InlineLimit:  32,
		RawLeaves:    false,
		RawLeavesSet: false,

		Chunker: "size-262144",
		Layout:  BalancedLayout,

		Pin:      false,
		OnlyHash: false,
		FsCache:  false,
		NoCopy:   false,

		Events:   nil,
		Silent:   false,
		Progress: false,
	}

	for _, opt := range opts {
		err := opt(options)
		if err != nil {
			return nil, cid.Prefix{}, err
		}
	}

	// nocopy -> rawblocks
	if options.NoCopy && !options.RawLeaves {
		// fixed?
		if options.RawLeavesSet {
			return nil, cid.Prefix{}, fmt.Errorf("nocopy option requires '--raw-leaves' to be enabled as well")
		}

		// No, satisfy mandatory constraint.
		options.RawLeaves = true
	}

	// (hash != "sha2-256") -> CIDv1
	if options.MhType != mh.SHA2_256 {
		switch options.CidVersion {
		case 0:
			return nil, cid.Prefix{}, errors.New("CIDv0 only supports sha2-256")
		case 1, -1:
			options.CidVersion = 1
		default:
			return nil, cid.Prefix{}, fmt.Errorf("unknown CID version: %d", options.CidVersion)
		}
	} else {
		if options.CidVersion < 0 {
			// Default to CIDv0
			options.CidVersion = 0
		}
	}

	// cidV1 -> raw blocks (by default)
	if options.CidVersion > 0 && !options.RawLeavesSet {
		options.RawLeaves = true
	}

	prefix, err := dag.PrefixForCidVersion(options.CidVersion)
	if err != nil {
		return nil, cid.Prefix{}, err
	}

	prefix.MhType = options.MhType
	prefix.MhLength = -1

	return options, prefix, nil
}

func UnixfsWriteOptions(opts ...UnixfsWriteOption) (*UnixfsWriteSettings, cid.Prefix, error) {
	options := &UnixfsWriteSettings{
		CidVersion: -1,
		MhType:     mh.SHA2_256,

		Offset:    0,
		Create:    false,
		Parents:   false,
		Truncate:  false,
		Count:     0,
		RawLeaves: false,
	}

	for _, opt := range opts {
		err := opt(options)
		if err != nil {
			return nil, cid.Prefix{}, err
		}
	}

	// (hash != "sha2-256") -> CIDv1
	if options.MhType != mh.SHA2_256 {
		switch options.CidVersion {
		case 0:
			return nil, cid.Prefix{}, errors.New("CIDv0 only supports sha2-256")
		case 1, -1:
			options.CidVersion = 1
		default:
			return nil, cid.Prefix{}, fmt.Errorf("unknown CID version: %d", options.CidVersion)
		}
	} else {
		if options.CidVersion < 0 {
			// Default to CIDv0
			options.CidVersion = 0
		}
	}

	// cidV1 -> raw blocks (by default)
	if options.CidVersion > 0 && !options.RawLeaves {
		options.RawLeaves = true
	}

	prefix, err := dag.PrefixForCidVersion(options.CidVersion)
	if err != nil {
		return nil, cid.Prefix{}, err
	}

	prefix.MhType = options.MhType
	prefix.MhLength = -1

	return options, prefix, nil
}

func UnixfsLsOptions(opts ...UnixfsLsOption) (*UnixfsLsSettings, error) {
	options := &UnixfsLsSettings{
		ResolveChildren: true,
	}

	for _, opt := range opts {
		err := opt(options)
		if err != nil {
			return nil, err
		}
	}

	return options, nil
}

func UnixfsStatOptions(opts ...UnixfsStatOption) (*UnixfsStatSettings, error) {
	options := &UnixfsStatSettings{
		Format:    "",
		Hash:      false,
		Size:      false,
		WithLocal: false,
	}

	for _, opt := range opts {
		err := opt(options)
		if err != nil {
			return nil, err
		}
	}

	return options, nil
}

func UnixfsReadOptions(opts ...UnixfsReadOption) (*UnixfsReadSettings, error) {
	options := &UnixfsReadSettings{
		Offset: 0,
		Count:  0,
	}

	for _, opt := range opts {
		err := opt(options)
		if err != nil {
			return nil, err
		}
	}

	return options, nil
}

func UnixfsCpOptions(opts ...UnixfsCpOption) (*UnixfsCpSettings, error) {
	options := &UnixfsCpSettings{
		Parents: false,
	}

	for _, opt := range opts {
		err := opt(options)
		if err != nil {
			return nil, err
		}
	}

	return options, nil
}

func UnixfsRmOptions(opts ...UnixfsRmOption) (*UnixfsRmSettings, error) {
	options := &UnixfsRmSettings{
		Recursive: false,
		Force:     false,
	}

	for _, opt := range opts {
		err := opt(options)
		if err != nil {
			return nil, err
		}
	}

	return options, nil
}

func UnixfsMkdirOptions(opts ...UnixfsMkdirOption) (*UnixfsMkdirSettings, cid.Prefix, error) {
	options := &UnixfsMkdirSettings{
		Parents:    false,
		CidVersion: -1,
		MhType:     mh.SHA2_256,
	}

	for _, opt := range opts {
		err := opt(options)
		if err != nil {
			return nil, cid.Prefix{}, err
		}
	}

	// (hash != "sha2-256") -> CIDv1
	if options.MhType != mh.SHA2_256 {
		switch options.CidVersion {
		case 0:
			return nil, cid.Prefix{}, errors.New("CIDv0 only supports sha2-256")
		case 1, -1:
			options.CidVersion = 1
		default:
			return nil, cid.Prefix{}, fmt.Errorf("unknown CID version: %d", options.CidVersion)
		}
	} else {
		if options.CidVersion < 0 {
			// Default to CIDv0
			options.CidVersion = 0
		}
	}

	prefix, err := dag.PrefixForCidVersion(options.CidVersion)
	if err != nil {
		return nil, cid.Prefix{}, err
	}

	prefix.MhType = options.MhType
	prefix.MhLength = -1

	return options, prefix, nil
}

type unixfsOpts struct{}

var Unixfs unixfsOpts

// CidVersion specifies which CID version to use. Defaults to 0 unless an option
// that depends on CIDv1 is passed.
func (unixfsOpts) CidVersion(version int) UnixfsAddOption {
	return func(settings *UnixfsAddSettings) error {
		settings.CidVersion = version
		return nil
	}
}

// Hash function to use. Implies CIDv1 if not set to sha2-256 (default).
//
// Table of functions is declared in https://github.com/multiformats/go-multihash/blob/master/multihash.go
func (unixfsOpts) Hash(mhtype uint64) UnixfsAddOption {
	return func(settings *UnixfsAddSettings) error {
		settings.MhType = mhtype
		return nil
	}
}

// RawLeaves specifies whether to use raw blocks for leaves (data nodes with no
// links) instead of wrapping them with unixfs structures.
func (unixfsOpts) RawLeaves(enable bool) UnixfsAddOption {
	return func(settings *UnixfsAddSettings) error {
		settings.RawLeaves = enable
		settings.RawLeavesSet = true
		return nil
	}
}

// Inline tells the adder to inline small blocks into CIDs
func (unixfsOpts) Inline(enable bool) UnixfsAddOption {
	return func(settings *UnixfsAddSettings) error {
		settings.Inline = enable
		return nil
	}
}

// InlineLimit sets the amount of bytes below which blocks will be encoded
// directly into CID instead of being stored and addressed by it's hash.
// Specifying this option won't enable block inlining. For that use `Inline`
// option. Default: 32 bytes
//
// Note that while there is no hard limit on the number of bytes, it should be
// kept at a reasonably low value, such as 64; implementations may choose to
// reject anything larger.
func (unixfsOpts) InlineLimit(limit int) UnixfsAddOption {
	return func(settings *UnixfsAddSettings) error {
		settings.InlineLimit = limit
		return nil
	}
}

// Chunker specifies settings for the chunking algorithm to use.
//
// Default: size-262144, formats:
// size-[bytes] - Simple chunker splitting data into blocks of n bytes
// rabin-[min]-[avg]-[max] - Rabin chunker
func (unixfsOpts) Chunker(chunker string) UnixfsAddOption {
	return func(settings *UnixfsAddSettings) error {
		settings.Chunker = chunker
		return nil
	}
}

// Layout tells the adder how to balance data between leaves.
// options.BalancedLayout is the default, it's optimized for static seekable
// files.
// options.TrickleLayout is optimized for streaming data,
func (unixfsOpts) Layout(layout Layout) UnixfsAddOption {
	return func(settings *UnixfsAddSettings) error {
		settings.Layout = layout
		return nil
	}
}

// Pin tells the adder to pin the file root recursively after adding
func (unixfsOpts) Pin(pin bool) UnixfsAddOption {
	return func(settings *UnixfsAddSettings) error {
		settings.Pin = pin
		return nil
	}
}

// HashOnly will make the adder calculate data hash without storing it in the
// blockstore or announcing it to the network
func (unixfsOpts) HashOnly(hashOnly bool) UnixfsAddOption {
	return func(settings *UnixfsAddSettings) error {
		settings.OnlyHash = hashOnly
		return nil
	}
}

// Events specifies channel which will be used to report events about ongoing
// Add operation.
//
// Note that if this channel blocks it may slowdown the adder
func (unixfsOpts) Events(sink chan<- interface{}) UnixfsAddOption {
	return func(settings *UnixfsAddSettings) error {
		settings.Events = sink
		return nil
	}
}

// Silent reduces event output
func (unixfsOpts) Silent(silent bool) UnixfsAddOption {
	return func(settings *UnixfsAddSettings) error {
		settings.Silent = silent
		return nil
	}
}

// Progress tells the adder whether to enable progress events
func (unixfsOpts) Progress(enable bool) UnixfsAddOption {
	return func(settings *UnixfsAddSettings) error {
		settings.Progress = enable
		return nil
	}
}

// ToFiles tells the adder whether to add reference to Files API (MFS) at the provided path
func (unixfsOpts) ToFiles(path string) UnixfsAddOption {
	return func(settings *UnixfsAddSettings) error {
		settings.ToFiles = path
		return nil
	}
}

// FsCache tells the adder to check the filestore for pre-existing blocks
//
// Experimental
func (unixfsOpts) FsCache(enable bool) UnixfsAddOption {
	return func(settings *UnixfsAddSettings) error {
		settings.FsCache = enable
		return nil
	}
}

// NoCopy tells the adder to add the files using filestore. Implies RawLeaves.
//
// Experimental
func (unixfsOpts) Nocopy(enable bool) UnixfsAddOption {
	return func(settings *UnixfsAddSettings) error {
		settings.NoCopy = enable
		return nil
	}
}

func (unixfsOpts) ResolveChildren(resolve bool) UnixfsLsOption {
	return func(settings *UnixfsLsSettings) error {
		settings.ResolveChildren = resolve
		return nil
	}
}

func (unixfsOpts) UseCumulativeSize(use bool) UnixfsLsOption {
	return func(settings *UnixfsLsSettings) error {
		settings.UseCumulativeSize = use
		return nil
	}
}

// Parents No error if existing, make parent directories as needed
func (unixfsOpts) Parents(parents bool) UnixfsMkdirOption {
	return func(settings *UnixfsMkdirSettings) error {
		settings.Parents = parents
		return nil
	}
}

// MkdirCidVersion cid version to use.
//
// Experimental
func (unixfsOpts) MkdirCidVersion(cidVer int) UnixfsMkdirOption {
	return func(settings *UnixfsMkdirSettings) error {
		settings.CidVersion = cidVer
		return nil
	}
}

// MkdirHash Hash function to use. Will set Cid version to 1 if used.
//
// Experimental
func (unixfsOpts) MkdirHash(mhtype uint64) UnixfsMkdirOption {
	return func(settings *UnixfsMkdirSettings) error {
		settings.MhType = mhtype
		return nil
	}
}

// Recursive remove directories
func (unixfsOpts) Recursive(recursive bool) UnixfsRmOption {
	return func(settings *UnixfsRmSettings) error {
		settings.Recursive = recursive
		return nil
	}
}

// Force forcibly remove target at path
func (unixfsOpts) Force(force bool) UnixfsRmOption {
	return func(settings *UnixfsRmSettings) error {
		settings.Force = force
		return nil
	}
}

// Parents make parent directories as needed
func (unixfsOpts) CpParents(parent bool) UnixfsCpOption {
	return func(settings *UnixfsCpSettings) error {
		settings.Parents = parent
		return nil
	}
}

// Parents make parent directories as needed
func (unixfsOpts) Offset(offset int64) UnixfsReadOption {
	return func(settings *UnixfsReadSettings) error {
		settings.Offset = offset
		return nil
	}
}

// Parents make parent directories as needed
func (unixfsOpts) Count(count int64) UnixfsReadOption {
	return func(settings *UnixfsReadSettings) error {
		settings.Count = count
		return nil
	}
}

// Format print statistics in given format
func (unixfsOpts) Format(format string) UnixfsStatOption {
	return func(settings *UnixfsStatSettings) error {
		settings.Format = format
		return nil
	}
}

// Hash print only hash
func (unixfsOpts) StatHash(hash bool) UnixfsStatOption {
	return func(settings *UnixfsStatSettings) error {
		settings.Hash = hash
		return nil
	}
}

// StatSize print only size
func (unixfsOpts) StatSize(size bool) UnixfsStatOption {
	return func(settings *UnixfsStatSettings) error {
		settings.Size = size
		return nil
	}
}

// WithLocal compute the amount of the dag that is local, and if possible the total size.
func (unixfsOpts) WithLocal(withLocal bool) UnixfsStatOption {
	return func(settings *UnixfsStatSettings) error {
		settings.WithLocal = withLocal
		return nil
	}
}

// WriteOffset
func (unixfsOpts) WriteOffset(offset int64) UnixfsWriteOption {
	return func(settings *UnixfsWriteSettings) error {
		settings.Offset = offset
		return nil
	}
}

// Create
func (unixfsOpts) Create(create bool) UnixfsWriteOption {
	return func(settings *UnixfsWriteSettings) error {
		settings.Create = create
		return nil
	}
}

// WriteParents
func (unixfsOpts) WriteParents(parents bool) UnixfsWriteOption {
	return func(settings *UnixfsWriteSettings) error {
		settings.Parents = parents
		return nil
	}
}

// Truncate
func (unixfsOpts) Truncate(truncate bool) UnixfsWriteOption {
	return func(settings *UnixfsWriteSettings) error {
		settings.Truncate = truncate
		return nil
	}
}

// WriteRawLeaves
func (unixfsOpts) WriteRawLeaves(rawLeaves bool) UnixfsWriteOption {
	return func(settings *UnixfsWriteSettings) error {
		settings.RawLeaves = rawLeaves
		return nil
	}
}

// WriteCount
func (unixfsOpts) WriteCount(count int64) UnixfsWriteOption {
	return func(settings *UnixfsWriteSettings) error {
		settings.Count = count
		return nil
	}
}

// WriteCidVersion
func (unixfsOpts) WriteCidVersion(cidVersion int) UnixfsWriteOption {
	return func(settings *UnixfsWriteSettings) error {
		settings.CidVersion = cidVersion
		return nil
	}
}

// WriteHash
func (unixfsOpts) WriteHash(mhtype uint64) UnixfsWriteOption {
	return func(settings *UnixfsWriteSettings) error {
		settings.MhType = mhtype
		return nil
	}
}
