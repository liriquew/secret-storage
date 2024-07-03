<script>
import KVList from "@/components/KVList.vue";
import BucketList from "@/components/BucketList.vue";

import Sidebar from "@/components/Sidebar.vue";
import HeaderTitle from "@/components/HeaderTitle.vue";
import Cookies from "js-cookie"

export default {
  components: {
    KVList,
    BucketList,
    Sidebar,
    HeaderTitle,
  },
  data() {
    return {
      addNewItem: false,
      kvs: [],
      buckets: [],
    }
  },
  getToken() {
    return ;
  },
  async created() {
    this.fetchData()
  },
  methods: {
    async fetchData() {
      fetch('http://localhost:8080/api/list/', {
        headers: {
          'Authorization': 'Bearer ' + Cookies.get('jwtToken'),
        }
      }).then(response => {
        if (!response.ok) {
          throw new Error('ERROR');
        }
        return response.json();
      }).then(data => {
        this.kvs = data.kvs;
        this.buckets = data.buckets
        console.log(data.kvs);
        console.log(data.buckets);
      }).catch(e => {
        console.log(e)
      })
    },
    addNewKV() {
      this.addNewItem = !this.addNewItem

    },
  }
}

</script>

<template>
  <HeaderTitle/>
  <Sidebar />
  <div class="main">
    <BucketList :buckets="buckets" />
    <div class="buttons">
      <button class="add-button" @click="addNewKV">Добавить запись</button>
    </div>
    <div v-if="addNewItem==true"> 
      <p class="text">Новая запись:</p>
      <KVList :kvs="[{key, value}]" />
    </div>
    <div class="list"> 
      <KVList :kvs="kvs" />
    </div>
  </div>
</template>

<style>
.main {
  margin-left: 300px;
  margin-right: 100px;
  padding-top: 20px;
  align-items: center;
}

.text {
  text-align: center;
  font-size: 20px;
  font-family: Arial, sans-serif;
}

.list {
  padding-top: 20px;
}

.buttons {
  display: flex;
}

.add-button {
  display: inline-block;
  margin-right: 10px;
  padding: 10px 20px;
  font-size: 16px;
  font-weight: bold;
  color: #ffffff;
  background-color: #007bff;
  border: none;
  border-radius: 5px;
  cursor: pointer;
  transition: background-color 0.3s ease;
}

.add-button:hover {
  background-color: #0056b3;
}
</style>