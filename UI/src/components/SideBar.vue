<script>
import Cookies from 'js-cookie';

export default {
  data() {
    return {
      username: "", 
    }
  },
  async created() {
    this.username = localStorage.getItem("username")
  },
  methods: {
    clearUser() {
      localStorage.setItem("username", null)
      Cookies.remove("jwtToken")
      this.$router.push("/")
    }
  }
}
</script>
<template>
  <div class="sidenav"> 
    <div class="title"> {{ this.username }} </div>
    <div class="title">
      <button class="bttn" @click="clearUser">Выйти</button>
    </div>
  </div>
</template>

<style>
.bttn {
  background-color: black;
  content: 'Удалить';
  color: white;
  padding: 10px 15px;
  border: 1px solid white;
  border-radius: 8px;
  z-index: 1000;
  font-size: 17px;
}

.title {
  width: 100%;
  padding-top: 20px;
  color: white;
  text-align: center;
  font-size: 25px;
  font-weight: bold;
  font-family: Arial, sans-serif;
}

.sidenav {
  position: fixed;
  width: 250px;
  height: 100%;
  padding-top: 100px;
  background-color: black;
  overflow-x: hidden;
  z-index: 0;
}
</style>