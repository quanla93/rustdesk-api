package web

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/global"
)

type Index struct {
}

func (i *Index) Index(c *gin.Context) {
	c.Redirect(302, "/_admin/")
}

func (i *Index) ConfigJs(c *gin.Context) {
	apiServer := global.Config.Rustdesk.ApiServer
	idServer := global.Config.Rustdesk.IdServer
	relayServer := global.Config.Rustdesk.RelayServer
	key := global.Config.Rustdesk.Key
	magicQueryonline := global.Config.Rustdesk.WebclientMagicQueryonline
	wsHost := global.Config.Rustdesk.WsHost

	customConfig := fmt.Sprintf(`[default-settings]
custom-rendezvous-server=%s
relay-server=%s
api-server=%s
key=%s
`, idServer, relayServer, apiServer, key)
	customConfigJSON, _ := json.Marshal(customConfig)
	apiServerJSON, _ := json.Marshal(apiServer)
	idServerJSON, _ := json.Marshal(idServer)
	relayServerJSON, _ := json.Marshal(relayServer)
	keyJSON, _ := json.Marshal(key)
	wsHostJSON, _ := json.Marshal(wsHost)

	tmp := fmt.Sprintf(`(function () {
  const apiServer = %s;
  const idServer = %s;
  const relayServer = %s;
  const key = %s;
  const customConfig = %s;
  const ws2Prefix = 'wc-';

  localStorage.setItem('api-server', apiServer);
  localStorage.setItem(ws2Prefix + 'api-server', apiServer);
  localStorage.setItem(ws2Prefix + 'custom-rendezvous-server', idServer);
  localStorage.setItem(ws2Prefix + 'relay-server', relayServer);
  localStorage.setItem(ws2Prefix + 'key', key);

  const customConfigNode = document.getElementById('custom-config');
  if (customConfigNode) {
    customConfigNode.textContent = btoa(customConfig);
  }

  window.webclient_magic_queryonline = %d;
  window.ws_host = %s;
})();
`, apiServerJSON, idServerJSON, relayServerJSON, keyJSON, customConfigJSON, magicQueryonline, wsHostJSON)

	c.Header("Content-Type", "application/javascript")
	c.String(200, tmp)
}
