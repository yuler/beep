: "${CORE_PORT:?CORE_PORT is required. Copy .env.example to .env.}"

echo "Using Cloudflare tunnel URL: http://localhost:${CORE_PORT}"
echo ""
echo ""
echo "=================================================="
echo "Go to https://www.creem.io/dashboard/developers/webhooks setup the webhook URL"
echo "=================================================="
echo ""
echo ""

cloudflared tunnel --url "http://localhost:${CORE_PORT}"
