<script>
import Cookies from 'js-cookie';

export default {
  props: {
    kv: {
      type: Object,
      required: true,
    }
  },
  data() {
    return {
      localKV: { ...this.kv }
    };
  },
  methods: {
    async saveKV() {
      var path = this.$route.path.replace('/keys/', '')   // /keys/{path}/.../keys/{path}
      if (path == "/keys") {
        path = ""
      }

      await fetch('http://localhost:8080/api/secrets/' + path, {
        method: "post",
        headers: {
          'Authorization': 'Bearer ' + Cookies.get('jwtToken'),
          "Content-Type": "application/json"
        },
        body: JSON.stringify({key:this.localKV.key, value:this.localKV.value})
      }).then(response => {
        if (!response.ok) {
          throw new Error('Network response was not ok ' + response.statusText);
        }
      }).catch(error => {
        console.error(error);
      })
    },

    deleteKV() {
      this.$emit('deleteKV', this.localKV.key);
    },
  }
}
</script>

<template>
  <div class="form-container">
    <div class="title-div">
      <label for="key">Key</label>
      <input type="text" id="key" v-model="localKV.key" readonly>
    </div>
    <div class="title-div">
      <label for="value">Value</label>
      <input type="text" id="value" v-model="localKV.value">
    </div>
    <div class="button-container">
      <button type="submit" class="submit" @click="saveKV"></button>
      <button type="reset" class="reset" @click="deleteKV "></button>
    </div>
  </div>
</template>

<style scoped>
.form-container {
  align-items: center;
  display: flex;
  justify-content: space-between;
  width: 100%;
  padding: 10px 20px;
  margin-top: 10px;
  border: 1px solid #ccc;
  border-radius: 8px;
}

.title-div {
  text-align: left;
  font-weight: bold;
  font-family: Arial, sans-serif;
}

.form-container label {
  font-size: 18px;
}

.form-container input {
  padding: 5px;
  font-size: 16px;
  width: 100%;
}

.form-container button {
  margin: 0 5px;
  padding: 5px 10px;
  font-size: 16px;
  border: none;
  background: none;
  cursor: pointer;
}

.form-container button.submit::before {
  content: 'Сохранить';
  color: blue;
  padding: 10px 15px;
  border: 1px solid blue;
  border-radius: 8px;
}

.form-container button.reset::before {
  content: 'Удалить';
  color: red;
  padding: 10px 15px;
  border: 1px solid red;
  border-radius: 8px;
}
</style>