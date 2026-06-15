import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import './style.css'

const app = createApp(App)
app.use(createPinia()) // pinia 先于 router（守卫里要用 store）
app.use(router)
app.mount('#app')
