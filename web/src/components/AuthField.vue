<script>
import Cookies from "js-cookie";
export default {
  data() {
    return {
      selectedAuthMethod: 'login',
      login: '',
      password: '',
      token: '',
    }
  },
  methods: {
    async authFetch() {
      let requestBody = {}
      if (this.selectedAuthMethod == 'token') {
        requestBody = {
          token: this.token
        }
      } else {
        requestBody = {
          username: this.login,
          password: this.password,
        }
      }
      fetch('http://localhost:8080/api/signin', {
        method: "post",
        headers: {
          "Content-Type": "application/json"
        },
        body: JSON.stringify(requestBody)
      }).then(response => {
        if (response.ok) {
          return response.json();
        }
        throw new Error('Network response was not ok ' + response.statusText);
      }).then(data => {
        const token = data.token;

        Cookies.set('jwtToken', token, { path: '/' });

        console.log('Token:', token);
        this.$router.push('keys');
      }).catch(error => {
        console.error('There has been a problem with your fetch operation:', error);
      })
    },

    inputToken(event) {
      this.token = event.target.value
    },

    inputLogin(event) {
      this.login = event.target.value
    },

    inputPassword(event) {
      this.password = event.target.value
    }
  }
}
</script>

<template>
  <div class="form_auth_block">
    <div class="form_auth_block_content">
      <p class="form_auth_block_head_text">Авторизация</p>

      <select class="form_auth_button" id="authMethod" v-model="selectedAuthMethod">
        <option value="token">Токен</option>
        <option value="login">Логин/пароль</option>
      </select>


      <form class="form_auth_style" @submit.prevent>

        <div v-if="selectedAuthMethod==='token'">
          <label>Токен пользователя</label>
          <input :value="token" @input="inputToken" required>
        </div>

        <div v-if="selectedAuthMethod=='login'">
          <label>Имя пользователя</label>
          <input :value="login" @input="inputLogin" required>
          <label>Пароль</label>
          <input :value="password" @input="inputPassword" required>
        </div>

        <button class="form_auth_button" @click="authFetch">Войти</button>
      </form>
    </div>
  </div>
</template>

<style scoped>
.form_auth_block{
  width: 500px;
  height: auto;
  padding-top: 30px;
  padding-bottom: 40px;
  margin: 30px auto;
  border: 1px solid #ccc;
  border-radius: 8px;
  font-family: Arial, sans-serif;
}


.form_auth_block_head_text{
  display: block;
  text-align: center;
  font-size: 20px;
  font-weight: 600;
  background: #ffffff;
  opacity: 0.7;
}

.form_auth_block label{
  display: block;
  text-align: center;
  padding-top: 20px;
  opacity: 0.7;
  margin-bottom: 10px;
  margin-top: 10px;
}

.form_auth_block input{
  display: block;
  margin: 0 auto;
  width: 80%;
  height: 45px;
  border: 1px solid #000000;
  border-radius: 5px;
}
input:focus {
  color: #000000;
  border: 1px solid #000000;
  border-radius: 5px;
}

.form_auth_button{
  display: block;
  width: 80%;
  margin: 30px auto 0;
  border: 1px solid #000000;
  border-radius: 5px;
  height: 35px;
  cursor: pointer;
  text-align: center;
  font-size: 16px;
}

</style>