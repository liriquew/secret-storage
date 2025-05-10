<script>
import HeaderTitle from "@/components/HeaderTitle.vue"
import config from "@/config"

export default {
    components: {
        HeaderTitle,
    },
    data() {
        return {
            keyPart: "",
        }
    },
    methods: {
        async handleSubmit(event) {
            event.preventDefault();

            const response = await fetch(`${config["api_host"]}/api/unseal?part=${this.keyPart}`, {
                method: "POST",
            })

            if (response.ok) {
                this.keyPart = ""
            }
        },
        async combineComplete() {
            const response = await fetch(`${config["api_host"]}/api/unseal/complete`, {method: "POST"})

            if (response.ok) {
                this.$router.push("/")
            }
        },
        inputPart(event) {
            this.keyPart = event.target.value
        }
    }
}
</script>

<template>
    <HeaderTitle />
    <div style="padding-top: 100px;">
        <div class="form_block">
            <p class="form_head_text">Разблокировка</p>

            <form @submit="handleSubmit">
                <label for="userText" class="form_label">Часть мастер-ключа:</label><br>
                <input class="form_textarea" :value="keyPart" @input="inputPart" name="userText" rows="10" cols="30">
                <button type="submit" class="form_button">Отправить</button>
            </form>

            <button class="form_button" style="margin-top: 18px;" @click="combineComplete">Все части отправлены<br> (в том числе и участников)</button>
        </div>
    </div>
</template>

<style scoped>
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

.form_textarea {
    display: block;
    width: 90%;
    margin: 0 auto 20px;
    border: 1px solid #ccc;
    border-radius: 5px;
    padding: 10px;
    font-size: 16px;
    font-family: Arial, sans-serif;
    box-sizing: border-box;
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
