package cache

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/crypto/ssh"
)

// Redis .redis cache
type Redis struct {
	ctx  context.Context
	conn redis.UniversalClient
}

// RedisOpts redis 连接属性
type RedisOpts struct {
	Host        string `yml:"host" json:"host"`
	Password    string `yml:"password" json:"password"`
	Database    int    `yml:"database" json:"database"`
	MaxIdle     int    `yml:"max_idle" json:"max_idle"`
	MaxActive   int    `yml:"max_active" json:"max_active"`
	IdleTimeout int    `yml:"idle_timeout" json:"idle_timeout"` // second
	Dialer      func(ctx context.Context, network, addr string) (net.Conn, error)
}

// NewRedis 实例化
func NewRedis(ctx context.Context, opts *RedisOpts) *Redis {
	conn := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:        []string{opts.Host},
		DB:           opts.Database,
		Password:     opts.Password,
		IdleTimeout:  time.Second * time.Duration(opts.IdleTimeout),
		MinIdleConns: opts.MaxIdle,
		Dialer:       opts.Dialer,
	})
	return &Redis{ctx: ctx, conn: conn}
}

// NewRedisOverSSH 实例化（通过 SSH 代理连接 Redis ）
func NewRedisOverSSH(ctx context.Context, opts *RedisOpts, overSSH *OverSSH) *Redis {
	conn := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:        []string{opts.Host},
		DB:           opts.Database,
		Password:     opts.Password,
		IdleTimeout:  time.Second * time.Duration(opts.IdleTimeout),
		MinIdleConns: opts.MaxIdle,
		Dialer:       overSSH.MakeDialer(),
	})
	return &Redis{ctx: ctx, conn: conn}
}

// SetConn 设置conn
func (r *Redis) SetConn(conn redis.UniversalClient) {
	r.conn = conn
}

// SetRedisCtx 设置redis ctx 参数
func (r *Redis) SetRedisCtx(ctx context.Context) {
	r.ctx = ctx
}

// Get 获取一个值
func (r *Redis) Get(key string) interface{} {
	return r.GetContext(r.ctx, key)
}

// GetContext 获取一个值
func (r *Redis) GetContext(ctx context.Context, key string) interface{} {
	result, err := r.conn.Do(ctx, "GET", key).Result()
	if err != nil {
		return nil
	}
	return result
}

// Set 设置一个值
func (r *Redis) Set(key string, val interface{}, timeout time.Duration) error {
	return r.SetContext(r.ctx, key, val, timeout)
}

// SetContext 设置一个值
func (r *Redis) SetContext(ctx context.Context, key string, val interface{}, timeout time.Duration) error {
	return r.conn.SetEX(ctx, key, val, timeout).Err()
}

// IsExist 判断key是否存在
func (r *Redis) IsExist(key string) bool {
	return r.IsExistContext(r.ctx, key)
}

// IsExistContext 判断key是否存在
func (r *Redis) IsExistContext(ctx context.Context, key string) bool {
	result, _ := r.conn.Exists(ctx, key).Result()

	return result > 0
}

// Delete 删除
func (r *Redis) Delete(key string) error {
	return r.DeleteContext(r.ctx, key)
}

// DeleteContext 删除
func (r *Redis) DeleteContext(ctx context.Context, key string) error {
	return r.conn.Del(ctx, key).Err()
}

// SSHAuthMethod SSH认证方式
type SSHAuthMethod uint8

const (
	// PubKeyAuth SSH公钥方式认证
	PubKeyAuth SSHAuthMethod = 1
	// PwdAuth SSH密码方式认证
	PwdAuth SSHAuthMethod = 2
)

type OverSSH struct {
	Host       string        `yml:"host" json:"host"`
	Port       int           `yml:"port" json:"port"`
	AuthMethod SSHAuthMethod `yml:"auth_method" json:"auth_method"`
	Username   string        `yml:"username" json:"username"`
	Password   string        `yml:"password" json:"password"`
	KeyFile    string        `yml:"key_file" json:"key_file"`
}

func (s *OverSSH) DialWithPassword() (*ssh.Client, error) {
	return ssh.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", s.Host, s.Port),
		&ssh.ClientConfig{
			User: s.Username,
			Auth: []ssh.AuthMethod{
				ssh.Password(s.Password),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		},
	)
}

func (s *OverSSH) DialWithKeyFile() (*ssh.Client, error) {
	k, err := os.ReadFile(s.KeyFile)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(k)
	if err != nil {
		return nil, err
	}

	return ssh.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", s.Host, s.Port),
		&ssh.ClientConfig{
			User: s.Username,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		},
	)
}

func (s *OverSSH) MakeDialer() func(ctx context.Context, network, addr string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		var err error
		var sshclient *ssh.Client
		switch s.AuthMethod {
		case PwdAuth:
			sshclient, err = s.DialWithPassword()
		case PubKeyAuth:
			sshclient, err = s.DialWithKeyFile()
		}
		if err != nil {
			return nil, err
		}
		return sshclient.Dial(network, addr)
	}
}
