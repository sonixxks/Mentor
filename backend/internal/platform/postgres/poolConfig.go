package postgres

type PoolConfig struct {
	ConnString string
	MaxConns   int32
}
