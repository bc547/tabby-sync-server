import store from './store.js'

export default {
    async getSyncTokens() {
        return apiRequest("/admin/api/1/synctokens", "GET")
            .then(data => {
                const list = []
                for (const token of data.synctokens) {
                    list.push({
                        synctoken: token
                    })
                }
                return list
            })
    },

    async createSyncToken() {
        return apiRequest("/admin/api/1/synctokens", "POST")
            .then(data => {
                return data.synctoken
            })
    },

    async deleteSyncToken(reqData) {
        return apiRequest("/admin/api/1/synctokens/" + reqData, "DELETE")
            .then(data => {
                return data
            })
    },
}

async function apiRequest(url, method, data) {
    const requestOptions = {
        method: method,
        headers: {
            "Content-Type": "application/json",
            "Authorization": "Bearer " + store.$adminToken,
        },
        body: JSON.stringify(data),
    }

    // fetch api examples: https://dmitripavlutin.com/javascript-fetch-async-await/
    // error handling: https://itnext.io/error-handling-with-async-await-in-js-26c3f20bc06a
    return await fetch(url, requestOptions).then(async response => {
        const resp = await response.json()

        if (!response.ok) {
            const message = `fetch error: ${method} ${url} -> ${response.status} ${response.statusText}`;
            throw new Error(message);
        }

        return resp
    })
}