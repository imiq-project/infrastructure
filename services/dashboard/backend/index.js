const express = require('express')
const path = require('path')
const { createProxyMiddleware } = require('http-proxy-middleware');

const app = express()
const port = 1000

dist_dir = path.resolve(__dirname + '/../frontend')
app.use(express.static(dist_dir))

app.disable('etag');

const orionEntitiesProxy = createProxyMiddleware({
  target: 'http://orion:1026/v2/entities',
  changeOrigin: true
});

// Only forward GET requests
app.use('/entities', (req, res, next) => {
  if (req.method === 'GET') {
    orionEntitiesProxy(req, res, next);
  } else {
    res.status(405).send('Method Not Allowed');
  }
});

app.get('/', (req, res) => {
    res.sendFile(dist_dir + '/index.html');
})

app.listen(port, () => {
    console.log(`Example app listening on port ${port}`)
})
