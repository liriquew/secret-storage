<script>
import KVList from "@/components/KVList.vue";
import BucketList from "@/components/BucketList.vue";
import NewKVItem from "@/components/NewKVItem.vue";

import Sidebar from "@/components/SideBar.vue";
import HeaderTitle from "@/components/HeaderTitle.vue";
import Cookies from "js-cookie"

export default {
  components: {
    KVList,
    BucketList,
    Sidebar,
    HeaderTitle,
    NewKVItem,
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
  watch: {
    '$route.params.path': 'fetchData'
  },
  async created() {
    const token = Cookies.get('jwtToken')
    if (token == null) {
      this.$router.push("/")
    }
    this.fetchData()
  },
  emits: ['SaveAndAddKV', 'deleteKV'],
  methods: {
    curPath () {
        var path = this.$route.path.replace('/keys/', '')   // /keys/{path}/.../keys/{path}
        if (path == "/keys") {
            path = ""
        }
        return path
    },

    addNewKV() {
      this.addNewItem = !this.addNewItem
    },

    async fetchData() {
      var path = this.$route.path.replace('/keys/', '')
      if (path == "/keys") {
        path = ""
      }

      fetch('http://localhost:8080/api/list/'+path, {
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
      }).catch(e => {
        console.log(e)
      })
    },

    async SaveAndAddKV(path, KV) {
      var currentPath = 'http://localhost:8080/api/secrets/'
      var relevatePath = this.curPath()

      if (relevatePath != "") {
        currentPath += relevatePath
        if (currentPath[currentPath.length - 1] != "/") {
          currentPath += "/"
        }
      }

      await fetch(currentPath + path, {
        method: "POST",
        headers: {
          'Authorization': 'Bearer ' + Cookies.get('jwtToken'),
          "Content-Type": "application/json"
        },
        body: JSON.stringify(KV)
      })
      this.addNewItem = false
      this.fetchData()
    },

    async deleteKV(key) {
      var currentPath = 'http://localhost:8080/api/secrets/'
      var relevatePath = this.curPath()

      if (relevatePath != "") {
        currentPath += relevatePath + "/"
      }

      const response = await fetch(currentPath + key, {
        method: "DELETE",
        headers: {
          'Authorization': 'Bearer ' + Cookies.get('jwtToken')
        },
      });

      if (!response.ok) {
        throw new Error('ERROR');
      }
      
      this.kvs = this.kvs.filter(item => item.key !== key);
      
      const data = await response.json();
      const deletedParts = data;

      if (deletedParts == 0) {
        return
      }

      const relPath = this.curPath()
      const relPathParts= relPath.split("/");

      if (relPathParts.length - deletedParts == 0) {
        this.$router.push("/keys")
        return
      }

      const joinedParts = relPathParts.slice(0, relPathParts.length - deletedParts).join("/");
      this.$router.push("/keys/" + joinedParts);
    }
  }
}

</script>

<template>
  <HeaderTitle/>
  <Sidebar />
  <div class="main">
    <div v-if="this.buckets != null">
      <BucketList :buckets="this.buckets" />
    </div>
    <div class="buttons">
      <button class="add-button" @click="addNewKV">Добавить запись</button>
    </div>
    <div v-if="addNewItem == true"> 
      <NewKVItem @SaveAndAddKV="SaveAndAddKV" />
    </div>
    <KVList @deleteKV="deleteKV" :kvs="this.kvs" />
  </div>
</template>

<style>
.main {
  margin-left: 300px;
  margin-right: 100px;
  padding-top: 90px;
  align-items: center;
}

.buttons {
  display: flex;
  margin-top: 10px;
  margin-bottom: 30px;
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