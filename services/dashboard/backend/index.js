const express = require('express')
const path = require('path')
const app = express()
const port = 1000

dist_dir = path.resolve(__dirname + '/../frontend/dist')

app.use('/assets', express.static(dist_dir + '/assets'))
app.disable('etag');
app.get('/', (req, res) => {
    res.sendFile(dist_dir + '/index.html');
})

app.get('/entities', async (req, res) => {
    let result = await fetch("http://orion:1026/v2/entities")
    res.send(await result.json())
})

app.listen(port, () => {
    console.log(`Example app listening on port ${port}`)
})
