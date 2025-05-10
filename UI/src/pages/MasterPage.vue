<script>
import HeaderTitle from "@/components/HeaderTitle.vue"

export default {
    components: {
        HeaderTitle,
    },
    data() {
        return {
            threshold: 2,
            socket: null,
            keyPart: "",
            panding: false,
        }
    },
    methods: {
        connect(event) {
            event.preventDefault();
            this.socket = new WebSocket(`http://localhost:8080/api/master?threshold=${this.threshold}`);
            this.panding = true;
            console.log(this.socket)
            this.socket.onopen = () => {
                console.log("WebSocket connected");
            };

            this.socket.onmessage = (event) => {
                console.log(event.data);
                this.keyPart = JSON.parse(event.data)["part"]
                console.log(this.keyPart)
            };

            this.socket.onerror = (error) => {
                console.error("WebSocket error:", error);
            };

            this.socket.onclose = () => {
                console.log("WebSocket disconnected");
            };
        },
        async masterComplete() {
            fetch("http://localhost:8080/api/master/complete", {method: "GET"});
            this.isVisible = true;
        },
        inputThreshold(event) {
            this.threshold = event.target.value;
        },
    }
}
</script>

<template>
    <HeaderTitle />
    <div style="padding-top: 100px;">
        <div class="form_block">
            <p class="form_head_text">Генерация мастер-ключа</p>

            <form @submit="connect">
                <label for="threshold" class="form_label">Пороговое значение</label>
                <input id="threshold" type="number" class="threshold" min="2" max="10" step="1" required 
                    :value="threshold" @input="inputThreshold">
                <button class="form_button" @click="connectSocket">Получить часть мастер-ключа</button>
            </form>

            <button class="form_button" style="margin-top: 20px;" @click="masterComplete">Завершить разделение секрета</button>

            <p class="form_label" style="padding-top:10px" v-if="panding">Ожидание завершения...</p>

            <input class="threshold" required readonly style="width: 80%; margin-top: 10px;" hidden
                :value="keyPart" v-if="keyPart.length != 0">
        </div>
    </div>
</template>

<style scoped>
.threshold {
  display: block;
  margin: 0 auto;
  margin-bottom: 10px;
  width: 60%;
  font-size: 16px;
  text-align: center;
  height: 45px;
  border: 1px solid #000000;
  border-radius: 5px;
}

.form_block {
    width: 500px;
    padding: 30px 20px 40px;
    margin: 30px auto;
    border: 1px solid #ccc;
    border-radius: 8px;
    font-family: Arial, sans-serif;
    background: #f9f9f9;
}

.form_head_text {
    text-align: center;
    font-size: 20px;
    font-weight: 600;
    margin-bottom: 20px;
}

.form_label {
    display: block;
    text-align: center;
    font-size: 16px;
    opacity: 0.7;
    margin-bottom: 10px;
}

.form_button {
    display: block;
    width: 80%;
    margin: 0 auto;
    padding: 10px;
    border: 1px solid #000;
    border-radius: 5px;
    background-color: #007bff;
    color: #fff;
    cursor: pointer;
    text-align: center;
    font-size: 16px;
}

.form_button:hover {
    background-color: #0056b3;
}
</style>
