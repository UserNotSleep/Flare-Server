const express = require("express");
const cors = require("cors");

const app = express();
app.use(cors());
app.use(express.json());

let messages = [];

app.get('/api/messages', (req,res) => {
    res.json(message);
});

app.post('/api/messages',(req,res)=>{
    const {text,senderName } = req.body;
    if(!text || !senderName){
        return res.status(400).json({ error:'Поля text и senderName обязательны'});
    }
    const newMessage = {
        id: Date.now().toString(),
        text: text,
        senderName: senderName,
        timestamp: Date.now()
    };
    messages.push(newMessage);
    res.status(201).json(newMessage);
});

const PORT = process.env.PORT || 3000;
app.listen(PORT, () => {
  console.log(`✅`);
});