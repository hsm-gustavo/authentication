package middlewares

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
)

const ClientInfoKey contextKey = "client_info"

type ClientInfo struct {
	IP netip.Addr
	UserAgent string
}

func ExtractClientInfo(next http.Handler, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ipAddr netip.Addr
		var err error
		
		userAgent := r.Header.Get("User-Agent")

		// tratando proxies como nginx ou cloudflare
		// tenta pegar ip do cloudflare primeiro
		rawIP := r.Header.Get("CF-Connecting-IP")

		// se não tiver cloudflare, tenta pegar do x-forwarded-for
		if rawIP == "" {
			rawIP = r.Header.Get("X-Forwarded-For")
		}

		if rawIP != "" {
			// se tiver proxy, tenta parsear
			ipAddr, err = netip.ParseAddr(rawIP)
			if err != nil {
				// alerta: pode ser tentativa de ip spoofing ou erro de configuraçao
				slog.Warn("IP inválido recebido do proxy", "raw_ip", rawIP, "user_agent", userAgent)
			}
		}

		// se o ip do proxy está invalido ou se o proxy nao enviou nenhum ip
		// fallback para remoteAddr
		if !ipAddr.IsValid() {
			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err == nil {
				ipAddr, _ = netip.ParseAddr(host)
			} else {
				// fallback extremo
				ipAddr, _ = netip.ParseAddr(r.RemoteAddr)
			}
		}		
		
		if !ipAddr.IsValid() {
			slog.Error("incapaz de determinar qualquer IP válido para a requisição", "remote_addr", r.RemoteAddr)
			http.Error(w, "Incapaz de identificar origem", http.StatusBadRequest)
			return
		}

		info := ClientInfo{
			IP: ipAddr,
			UserAgent: userAgent,
		}

		ctx := context.WithValue(r.Context(), ClientInfoKey, info)
		next.ServeHTTP(w, r.WithContext(ctx))

	})
}