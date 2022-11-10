package rcmgr

// DefaultLimitConfig is a struct for configuring default limits.
type DefaultLimitConfig struct {
	SystemBaseLimit BaseLimit
	SystemMemory    MemoryLimit

	TransientBaseLimit BaseLimit
	TransientMemory    MemoryLimit

	ServiceBaseLimit BaseLimit
	ServiceMemory    MemoryLimit

	ServicePeerBaseLimit BaseLimit
	ServicePeerMemory    MemoryLimit

	ProtocolBaseLimit BaseLimit
	ProtocolMemory    MemoryLimit

	ProtocolPeerBaseLimit BaseLimit
	ProtocolPeerMemory    MemoryLimit

	PeerBaseLimit BaseLimit
	PeerMemory    MemoryLimit

	ConnBaseLimit BaseLimit
	ConnMemory    int64

	StreamBaseLimit BaseLimit
	StreamMemory    int64
}

func (cfg *DefaultLimitConfig) WithSystemMemory(memFraction float64, minMemory, maxMemory int64) DefaultLimitConfig {
	refactor := memFraction / cfg.SystemMemory.MemoryFraction
	r := *cfg
	r.SystemMemory.MemoryFraction = memFraction
	r.SystemMemory.MinMemory = minMemory
	r.SystemMemory.MaxMemory = maxMemory
	r.TransientMemory.MemoryFraction *= refactor
	r.ServiceMemory.MemoryFraction *= refactor
	r.ServicePeerMemory.MemoryFraction *= refactor
	r.ProtocolMemory.MemoryFraction *= refactor
	r.ProtocolPeerMemory.MemoryFraction *= refactor
	r.PeerMemory.MemoryFraction *= refactor
	return r
}

// DefaultLimits are the limits used by the default limiter constructors.
var DefaultLimits = DefaultLimitConfig{
	SystemBaseLimit: BaseLimit{
		StreamsInbound:  4096 << 2,
		StreamsOutbound: 16384 << 2,
		Streams:         16384 << 2,
		ConnsInbound:    256 << 2,
		ConnsOutbound:   1024 << 2,
		Conns:           1024 << 2,
		FD:              512 << 2,
	},

	SystemMemory: MemoryLimit{
		MemoryFraction: 0.125,
		MinMemory:      128 << 22,
		MaxMemory:      1 << 32,
	},

	TransientBaseLimit: BaseLimit{
		StreamsInbound:  128 << 2,
		StreamsOutbound: 512 << 2,
		Streams:         512 << 2,
		ConnsInbound:    32 << 2,
		ConnsOutbound:   128 << 2,
		Conns:           128 << 2,
		FD:              128 << 2,
	},

	TransientMemory: MemoryLimit{
		MemoryFraction: 1,
		MinMemory:      64 << 22,
		MaxMemory:      64 << 22,
	},

	ServiceBaseLimit: BaseLimit{
		StreamsInbound:  2048 << 2,
		StreamsOutbound: 8192 << 2,
		Streams:         8192 << 2,
	},

	ServiceMemory: MemoryLimit{
		MemoryFraction: 0.125 / 4,
		MinMemory:      64 << 22,
		MaxMemory:      256 << 22,
	},

	ServicePeerBaseLimit: BaseLimit{
		StreamsInbound:  256 << 2,
		StreamsOutbound: 512 << 2,
		Streams:         512 << 2,
	},

	ServicePeerMemory: MemoryLimit{
		MemoryFraction: 0.125 / 16,
		MinMemory:      16 << 22,
		MaxMemory:      64 << 22,
	},

	ProtocolBaseLimit: BaseLimit{
		StreamsInbound:  1024 << 2,
		StreamsOutbound: 4096 << 2,
		Streams:         4096 << 2,
	},

	ProtocolMemory: MemoryLimit{
		MemoryFraction: 0.125 / 8,
		MinMemory:      64 << 22,
		MaxMemory:      128 << 22,
	},

	ProtocolPeerBaseLimit: BaseLimit{
		StreamsInbound:  128 << 2,
		StreamsOutbound: 256 << 2,
		Streams:         512 << 2,
	},

	ProtocolPeerMemory: MemoryLimit{
		MemoryFraction: 0.125 / 16,
		MinMemory:      16 << 22,
		MaxMemory:      64 << 22,
	},

	PeerBaseLimit: BaseLimit{
		StreamsInbound:  512 << 2,
		StreamsOutbound: 1024 << 2,
		Streams:         1024 << 2,
		ConnsInbound:    8 << 2,
		ConnsOutbound:   16 << 2,
		Conns:           16 << 2,
		FD:              8 << 2,
	},

	PeerMemory: MemoryLimit{
		MemoryFraction: 0.125 / 16,
		MinMemory:      64 << 22,
		MaxMemory:      128 << 22,
	},

	ConnBaseLimit: BaseLimit{
		ConnsInbound:  1 << 2,
		ConnsOutbound: 1 << 2,
		Conns:         1 << 2,
		FD:            1 << 2,
	},

	ConnMemory: 1 << 22,

	StreamBaseLimit: BaseLimit{
		StreamsInbound:  1 << 2,
		StreamsOutbound: 1 << 2,
		Streams:         1 << 2,
	},

	StreamMemory: 16 << 22,
}
