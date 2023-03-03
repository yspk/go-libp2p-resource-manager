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
		StreamsInbound:  4096 << 4,
		StreamsOutbound: 16384 << 4,
		Streams:         16384 << 4,
		ConnsInbound:    256 << 6,
		ConnsOutbound:   1024 << 4,
		Conns:           1024 << 4,
		FD:              512 << 4,
	},

	SystemMemory: MemoryLimit{
		MemoryFraction: 0.125,
		MinMemory:      128 << 22,
		MaxMemory:      1 << 34,
	},

	TransientBaseLimit: BaseLimit{
		StreamsInbound:  128 << 4,
		StreamsOutbound: 512 << 4,
		Streams:         512 << 4,
		ConnsInbound:    32 << 4,
		ConnsOutbound:   128 << 4,
		Conns:           128 << 4,
		FD:              128 << 4,
	},

	TransientMemory: MemoryLimit{
		MemoryFraction: 1,
		MinMemory:      64 << 20,
		MaxMemory:      64 << 24,
	},

	ServiceBaseLimit: BaseLimit{
		StreamsInbound:  2048 << 4,
		StreamsOutbound: 8192 << 4,
		Streams:         8192 << 4,
	},

	ServiceMemory: MemoryLimit{
		MemoryFraction: 0.125 / 4,
		MinMemory:      64 << 20,
		MaxMemory:      256 << 24,
	},

	ServicePeerBaseLimit: BaseLimit{
		StreamsInbound:  256 << 4,
		StreamsOutbound: 512 << 4,
		Streams:         512 << 4,
	},

	ServicePeerMemory: MemoryLimit{
		MemoryFraction: 0.125 / 16,
		MinMemory:      16 << 20,
		MaxMemory:      64 << 24,
	},

	ProtocolBaseLimit: BaseLimit{
		StreamsInbound:  1024 << 4,
		StreamsOutbound: 4096 << 4,
		Streams:         4096 << 4,
	},

	ProtocolMemory: MemoryLimit{
		MemoryFraction: 0.125 / 8,
		MinMemory:      64 << 20,
		MaxMemory:      128 << 24,
	},

	ProtocolPeerBaseLimit: BaseLimit{
		StreamsInbound:  128 << 4,
		StreamsOutbound: 256 << 4,
		Streams:         512 << 4,
	},

	ProtocolPeerMemory: MemoryLimit{
		MemoryFraction: 0.125 / 16,
		MinMemory:      16 << 20,
		MaxMemory:      64 << 24,
	},

	PeerBaseLimit: BaseLimit{
		StreamsInbound:  512 << 4,
		StreamsOutbound: 1024 << 4,
		Streams:         1024 << 4,
		ConnsInbound:    8 << 4,
		ConnsOutbound:   16 << 4,
		Conns:           16 << 4,
		FD:              8 << 4,
	},

	PeerMemory: MemoryLimit{
		MemoryFraction: 0.125 / 16,
		MinMemory:      64 << 20,
		MaxMemory:      128 << 24,
	},

	ConnBaseLimit: BaseLimit{
		ConnsInbound:  1 << 4,
		ConnsOutbound: 1 << 4,
		Conns:         1 << 4,
		FD:            1 << 4,
	},

	ConnMemory: 1 << 24,

	StreamBaseLimit: BaseLimit{
		StreamsInbound:  1 << 4,
		StreamsOutbound: 1 << 4,
		Streams:         1 << 4,
	},

	StreamMemory: 16 << 24,
}
