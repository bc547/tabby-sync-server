import * as pages from './pages/index.js'
import store from './store.js'

export default {
    name: 'App',
    components: Object.assign(pages),
    // components: Object.assign({homepage}, pages),

    setup() {
        const {watchEffect, onMounted, ref} = Vue;
        const page = ref(null);
        const menuIndex = ref(0);

        const menuItems = ref([
            {key: "page1", label: 'Info', icon: 'pi pi-info-circle', command: () => page.value = "page1"},
            {key: "page2", label: 'Admin', icon: 'pi pi-user-edit', command: () => page.value = "page2"},
        ]);

        //store management: save $variables to localstorage
        onMounted(() => {
            window.addEventListener('beforeunload', () => {
                Object.keys(store).forEach(function (key) {
                    if (key.charAt(0) == "$") {
                        localStorage.setItem(key, store[key]);
                    } else {
                        localStorage.removeItem("$" + key);
                    }
                });
            });
            Object.keys(store).forEach(function (key) {
                if (key.charAt(0) == "$" && localStorage.getItem(key)) {
                    store[key] = localStorage.getItem(key);
                }
            })
        })

        //url management
        watchEffect(() => {
            const match = /^(.*?)\/?(page\d+)?$/.exec(window.location.pathname)
            const url_path = match[1]
            let url_page = match[2] ? match[2] : 'page1'

            if (!page.value) {
                page.value = (url_page) ? url_page : 'page1'
            }

            window.history.pushState({page: page.value}, null, `${url_path}/${page.value}`)

            // select correct menu item
            for (const [i, item] of menuItems.value.entries()) {
                if (page.value === item.key) {
                    menuIndex.value = i
                    break
                }
            }

            window.onpopstate = function (event) {
                page.value = event.state.page
            };
        })

        return {page, pages, menuItems, menuIndex}
    },

    template: `<ConfirmDialog></ConfirmDialog>
<Toast position="top-center"></Toast>

<div>
    <TabMenu :model="menuItems" :activeIndex="menuIndex"/>
</div>
<div id="content">
    <component :is="page"></component>
</div>`,
};