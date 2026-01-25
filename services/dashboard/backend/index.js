const express = require('express')
const path = require('path')

const app = express()
const port = 1000

dist_dir = path.resolve(__dirname + '/../frontend')
app.use(express.static(dist_dir))

app.disable('etag');

app.get('/', (req, res) => {
    res.sendFile(dist_dir + '/index.html');
})

app.listen(port, () => {
    console.log(`Dashboard listening on port ${port}`)
})
