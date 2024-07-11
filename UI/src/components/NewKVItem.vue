<script>
export default {
  data() {
    return {
        key: "",
        value: "",
        path: "",
    };
  },
  methods: {
    saveKV() {
      this.$emit('SaveAndAddKV', this.path, {key:this.key, value:this.value});
    },
  }
}
</script>

<template>
    <div class="border">
        <p class="text">Новая запись:</p>
        <div class="form-container">
            <div class="title-div">
                <label for="path">Path</label>
                <input type="text" id="path" v-model="this.path">
            </div>
            <div style="display: flex; gap: 40px">
                <div class="title-div">
                    <label for="key">Key</label>
                    <input type="text" id="key" v-model="this.key">
                </div>
                <div class="title-div">
                    <label for="value">Value</label>
                    <input type="text" id="value" v-model="this.value">
                </div>
            </div>
            <div class="button-container">
                <button type="submit" class="submit" @click="saveKV"></button>
            </div>
        </div>
    </div>
</template>

<style scoped>

.path-input {
  margin: 10px 0;
  font-size: 20px
  
}

.border {
  width: 100%;
  padding: 10px 20px;
  margin-top: 10px;
  margin-bottom: 30px;
  border: 1px solid #ccc;
  border-radius: 8px;
}

.text {
  text-align: center;
  font-size: 20px;
  font-family: Arial, sans-serif;
  padding: 15px;
}

.form-container {
  align-items: center;
  display: flex;
  justify-content: space-between;

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
  padding: 10px 76px;
  border: 1px solid blue;
  border-radius: 8px;
}
</style>