export default {
    name: 'WebhookManager',
    data() {
        return {
            loading: false,
            webhooks: [],
            showAddForm: false,
            editingWebhookId: null,
            allEvents: [],
            expandedCategories: {
                connection: true,
                message: false,
                group: false,
                user: false,
                other: false
            },
            webhookForm: {
                url: '',
                secret: '',
                events: [],
                enabled: true,
                description: ''
            }
        }
    },
    methods: {
        async fetchAvailableEvents() {
            try {
                const response = await window.http.get('/webhook/events')
                this.allEvents = response.data.results || []
            } catch (error) {
                console.error('Failed to fetch available events:', error)
                showErrorInfo('Failed to load available events')
            }
        },
        openModal() {
            this.fetchWebhooks()
            this.fetchAvailableEvents()
            this.showAddForm = false
            $('#modalWebhookManager').modal({
                onApprove: function () {
                    return false;
                }
            }).modal('show');
        },
        async fetchWebhooks() {
            try {
                const response = await window.http.get('/webhook')
                this.webhooks = response.data.results || []
            } catch (error) {
                console.error('Failed to fetch webhooks:', error)
                showErrorInfo('Failed to load webhooks')
            }
        },
        isValidForm() {
            if (!this.webhookForm.url || !this.webhookForm.url.trim()) {
                return false
            }
            
            if (this.webhookForm.events.length === 0) {
                return false
            }
            
            return true
        },
        toggleEvent(event) {
            const index = this.webhookForm.events.indexOf(event)
            if (index > -1) {
                this.webhookForm.events.splice(index, 1)
            } else {
                this.webhookForm.events.push(event)
            }
        },
        isEventSelected(event) {
            return this.webhookForm.events.includes(event)
        },
        getSelectedEventsText() {
            if (this.webhookForm.events.length === 0) {
                return 'None'
            }
            return this.webhookForm.events.join(', ')
        },
        async handleSubmit() {
            if (!this.isValidForm() || this.loading) {
                return;
            }
            try {
                if (this.editingWebhookId) {
                    await this.updateApi()
                    showSuccessInfo('Webhook updated successfully')
                } else {
                    await this.submitApi()
                    showSuccessInfo('Webhook created successfully')
                }
                this.handleReset()
                this.showAddForm = false
                this.editingWebhookId = null
                await this.fetchWebhooks()
            } catch (err) {
                showErrorInfo(err)
            }
        },
        async submitApi() {
            this.loading = true;
            try {
                await window.http.post('/webhook', this.webhookForm)
            } catch (error) {
                if (error.response) {
                    throw new Error(error.response.data.message);
                }
                throw new Error(error.message);
            } finally {
                this.loading = false;
            }
        },
        handleReset() {
            this.webhookForm = {
                url: '',
                secret: '',
                events: [],
                enabled: true,
                description: ''
            }
            this.editingWebhookId = null
            // Reset expanded categories to default state
            this.expandedCategories = {
                connection: true,
                message: false,
                group: false,
                user: false,
                other: false
            }
        },

        async updateApi() {
            this.loading = true;
            try {
                await window.http.put(`/webhook/${this.editingWebhookId}`, this.webhookForm)
            } catch (error) {
                if (error.response) {
                    throw new Error(error.response.data.message);
                }
                throw new Error(error.message);
            } finally {
                this.loading = false;
            }
        },

        editWebhook(webhook) {
            this.webhookForm = {
                url: webhook.url,
                secret: webhook.secret || '',
                events: [...webhook.events],
                enabled: webhook.enabled,
                description: webhook.description || ''
            }
            this.editingWebhookId = webhook.id
            this.showAddForm = true
            
            this.$nextTick(() => {
                if (this.$refs.urlInput) {
                    this.$refs.urlInput.focus()
                }
            })
        },

        toggleAddForm() {
            this.showAddForm = !this.showAddForm
            if (!this.showAddForm) {
                this.handleReset()
            } else {
                this.$nextTick(() => {
                    if (this.$refs.urlInput) {
                        this.$refs.urlInput.focus()
                    }
                    this.initializeCheckboxes()
                })
            }
        },
        toggleCategory(category) {
            this.expandedCategories[category] = !this.expandedCategories[category]
            if (this.expandedCategories[category]) {
                this.initializeCheckboxes()
            }
        },
        selectAllEvents() {
            this.webhookForm.events = [...this.allEvents]
        },
        deselectAllEvents() {
            this.webhookForm.events = []
        },
        async deleteWebhook(id) {
            if (!confirm('Are you sure you want to delete this webhook?')) {
                return
            }
            
            try {
                await window.http.delete(`/webhook/${id}`)
                showSuccessInfo('Webhook deleted successfully')
                await this.fetchWebhooks()
            } catch (error) {
                showErrorInfo('Failed to delete webhook')
            }
        },
        
        async toggleWebhookEnabled(webhook) {
            try {
                const updatedWebhook = {
                    ...webhook,
                    enabled: !webhook.enabled
                };
                
                await window.http.put(`/webhook/${webhook.id}`, updatedWebhook)
                showSuccessInfo(`Webhook ${updatedWebhook.enabled ? 'enabled' : 'disabled'} successfully`)
                await this.fetchWebhooks()
            } catch (error) {
                showErrorInfo('Failed to update webhook status')
                await this.fetchWebhooks()
            }
        },
        
        getEventsByCategory(category) {
            const eventCategories = {
                connection: ['qr', 'pair.success', 'pair.error', 'qr.scanned.without.multidevice', 'connected', 'keepalive.timeout', 'keepalive.restored', 'logged.out', 'stream.replaced', 'manual.login.reconnect', 'temporary.ban', 'connect.failure', 'client.outdated', 'cat.refresh.error', 'stream.error', 'disconnected'],
                message: ['message', 'message.ack', 'fb.message', 'undecryptable.message', 'history.sync', 'media.retry', 'receipt.delivered', 'receipt.read', 'receipt.read.self', 'receipt.played', 'message.delete', 'message.revoke'],
                group: ['group', 'group.join', 'group.leave', 'group.promote', 'group.demote', 'group.info', 'group.picture'],
                user: ['user.about', 'user.picture', 'identity.change', 'privacy.settings', 'presence', 'chat.presence'],
                other: ['blocklist', 'newsletter.join', 'newsletter.leave', 'newsletter.mute.change', 'newsletter.live.update', 'offline.sync.preview', 'offline.sync.completed']
            };
            
            return this.allEvents.filter(event => {
                if (eventCategories[category]) {
                    return eventCategories[category].includes(event);
                }
                return false;
            });
        },
        
        formatEventName(event) {
            return event.split('.').map(word => 
                word.charAt(0).toUpperCase() + word.slice(1)
            ).join(' ');
        },
        
        initializeCheckboxes() {
            // Initialize Semantic UI checkboxes after DOM updates
            this.$nextTick(() => {
                $('.ui.checkbox').checkbox()
            })
        }
    },
    template: `
    <div class="green card" @click="openModal" style="cursor: pointer">
        <div class="content">
            <a class="ui teal right ribbon label">App</a>
            <div class="header">Manage Webhooks</div>
            <div class="description">
                Configure webhook endpoints for event notifications
            </div>
        </div>
    </div>
    
    <!-- Webhook Manager Modal -->
    <div class="ui modal" id="modalWebhookManager">
        <i class="close icon"></i>
        <div class="header">
            Manage Webhooks
        </div>
        <div class="scrolling content">
            <transition name="fade" mode="out-in">
                <div v-if="!showAddForm" key="webhook-list">
                    <div class="ui segments" v-if="webhooks.length > 0">
                        <div class="ui segment" v-for="webhook in webhooks" :key="webhook.id">
                            <div class="ui grid middle aligned">
                                <div class="twelve wide column">
                                    <div class="ui small header" style="margin-bottom: 0.5rem;">
                                        <i class="globe icon"></i>
                                        {{ webhook.url }}
                                    </div>
                                    <div class="ui list horizontal" style="margin: 0.5rem 0;">
                                        <div class="toggle-switch-container" style="margin: 0.5rem 0;">
                                            <label class="toggle-switch">
                                                <input type="checkbox" :checked="webhook.enabled" @change="toggleWebhookEnabled(webhook)">
                                                <span class="toggle-slider"></span>
                                            </label>
                                            <span class="toggle-label">{{ webhook.enabled ? 'Active' : 'Inactive' }}</span>
                                        </div>
                                    </div>
                                    <div>
                                        <div class="event-cards">
                                            <div v-for="event in allEvents" :key="event"
                                                 class="event-card"
                                                 :class="{'active-event': webhook.events.includes(event), 'inactive-event': !webhook.events.includes(event)}">
                                                {{ event }}
                                            </div>
                                        </div>
                                    </div>
                                    <p v-if="webhook.description" class="ui description text" style="margin-top: 0.5rem; color: #666;">
                                        <i class="file alternate icon"></i>{{ webhook.description }}
                                    </p>
                                </div>
                                <div class="four wide column right aligned">
                                    <div class="ui buttons vertical">
                                        <button class="ui blue compact icon button" style="margin-bottom: 0.5rem;" @click="editWebhook(webhook)" title="Edit webhook">
                                            <i class="edit icon"></i>
                                        </button>
                                        <button class="ui red compact icon button" @click="deleteWebhook(webhook.id)" title="Delete webhook">
                                            <i class="trash alternate icon"></i>
                                        </button>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                    
                    <div class="ui message info" v-else>
                        <div class="header">
                            <i class="info circle icon"></i>
                            No webhooks configured
                        </div>
                        <p>You haven't configured any webhooks yet.</p>
                    </div>
                    
                    <div class="ui divider"></div>
                    
                    <div class="ui form">
                        <div class="field">
                            <button class="ui primary fluid button" @click="toggleAddForm">
                                <i class="plus icon"></i>
                                Add New Webhook
                            </button>
                        </div>
                    </div>
                </div>
                
                <div v-else key="add-form">
                    <div class="ui form">
                        <h5 class="ui dividing header">
                            {{ editingWebhookId ? 'Edit Webhook' : 'Create New Webhook' }}
                        </h5>
                        
                        <div class="field">
                            <label>Webhook URL</label>
                            <input v-model="webhookForm.url" type="url"
                                   placeholder="https://your-webhook-endpoint.com/callback"
                                   aria-label="Webhook URL" ref="urlInput">
                        </div>
                        
                        <div class="field">
                            <label>Secret Key (optional)</label>
                            <input v-model="webhookForm.secret" type="text"
                                   placeholder="Secret for HMAC signature"
                                   aria-label="Secret Key">
                            <div class="ui pointing label">
                                Leave empty to use default secret
                            </div>
                        </div>
                        
                        <div class="field">
                            <label>Events to Receive</label>
                            <div class="ui pointing below label">
                                Select the events you want to receive webhook notifications for
                            </div>
                            
                            <!-- Event categories for better organization -->
                            <div class="ui segments">
                                <!-- Connection Events -->
                                <div class="ui segment">
                                    <div class="category-header" @click="toggleCategory('connection')" style="cursor: pointer; display: flex; align-items: center; justify-content: space-between; padding: 0.5rem 0;">
                                        <div style="display: flex; align-items: center; gap: 0.5rem;">
                                            <i :class="['small icon', expandedCategories.connection ? 'caret down' : 'caret right']" style="margin: 0; font-size: 0.9em;"></i>
                                            <span class="ui blue text" style="font-weight: 600; font-size: 1.1em;">Connection Events</span>
                                        </div>
                                        <span class="ui circular mini label blue">{{ getEventsByCategory('connection').length }}</span>
                                    </div>
                                    <div v-show="expandedCategories.connection" class="ui three column grid" style="margin-top: 0.75rem; padding-left: 1.5rem;">
                                        <div class="column" v-for="event in getEventsByCategory('connection')" :key="event">
                                            <div class="ui checkbox" style="margin-bottom: 0.5rem;">
                                                <input type="checkbox" :id="'event-' + event"
                                                       :value="event" v-model="webhookForm.events">
                                                <label :for="'event-' + event" style="font-size: 0.9em;">{{ formatEventName(event) }}</label>
                                            </div>
                                        </div>
                                    </div>
                                </div>

                                <!-- Message Events -->
                                <div class="ui segment">
                                    <div class="category-header" @click="toggleCategory('message')" style="cursor: pointer; display: flex; align-items: center; justify-content: space-between; padding: 0.5rem 0;">
                                        <div style="display: flex; align-items: center; gap: 0.5rem;">
                                            <i :class="['small icon', expandedCategories.message ? 'caret down' : 'caret right']" style="margin: 0; font-size: 0.9em;"></i>
                                            <span class="ui green text" style="font-weight: 600; font-size: 1.1em;">Message Events</span>
                                        </div>
                                        <span class="ui circular mini label green">{{ getEventsByCategory('message').length }}</span>
                                    </div>
                                    <div v-show="expandedCategories.message" class="ui three column grid" style="margin-top: 0.75rem; padding-left: 1.5rem;">
                                        <div class="column" v-for="event in getEventsByCategory('message')" :key="event">
                                            <div class="ui checkbox" style="margin-bottom: 0.5rem;">
                                                <input type="checkbox" :id="'event-' + event"
                                                       :value="event" v-model="webhookForm.events">
                                                <label :for="'event-' + event" style="font-size: 0.9em;">{{ formatEventName(event) }}</label>
                                            </div>
                                        </div>
                                    </div>
                                </div>

                                <!-- Group Events -->
                                <div class="ui segment">
                                    <div class="category-header" @click="toggleCategory('group')" style="cursor: pointer; display: flex; align-items: center; justify-content: space-between; padding: 0.5rem 0;">
                                        <div style="display: flex; align-items: center; gap: 0.5rem;">
                                            <i :class="['small icon', expandedCategories.group ? 'caret down' : 'caret right']" style="margin: 0; font-size: 0.9em;"></i>
                                            <span class="ui purple text" style="font-weight: 600; font-size: 1.1em;">Group Events</span>
                                        </div>
                                        <span class="ui circular mini label purple">{{ getEventsByCategory('group').length }}</span>
                                    </div>
                                    <div v-show="expandedCategories.group" class="ui three column grid" style="margin-top: 0.75rem; padding-left: 1.5rem;">
                                        <div class="column" v-for="event in getEventsByCategory('group')" :key="event">
                                            <div class="ui checkbox" style="margin-bottom: 0.5rem;">
                                                <input type="checkbox" :id="'event-' + event"
                                                       :value="event" v-model="webhookForm.events">
                                                <label :for="'event-' + event" style="font-size: 0.9em;">{{ formatEventName(event) }}</label>
                                            </div>
                                        </div>
                                    </div>
                                </div>

                                <!-- User Events -->
                                <div class="ui segment">
                                    <div class="category-header" @click="toggleCategory('user')" style="cursor: pointer; display: flex; align-items: center; justify-content: space-between; padding: 0.5rem 0;">
                                        <div style="display: flex; align-items: center; gap: 0.5rem;">
                                            <i :class="['small icon', expandedCategories.user ? 'caret down' : 'caret right']" style="margin: 0; font-size: 0.9em;"></i>
                                            <span class="ui orange text" style="font-weight: 600; font-size: 1.1em;">User Events</span>
                                        </div>
                                        <span class="ui circular mini label orange">{{ getEventsByCategory('user').length }}</span>
                                    </div>
                                    <div v-show="expandedCategories.user" class="ui three column grid" style="margin-top: 0.75rem; padding-left: 1.5rem;">
                                        <div class="column" v-for="event in getEventsByCategory('user')" :key="event">
                                            <div class="ui checkbox" style="margin-bottom: 0.5rem;">
                                                <input type="checkbox" :id="'event-' + event"
                                                       :value="event" v-model="webhookForm.events">
                                                <label :for="'event-' + event" style="font-size: 0.9em;">{{ formatEventName(event) }}</label>
                                            </div>
                                        </div>
                                    </div>
                                </div>

                                <!-- Other Events -->
                                <div class="ui segment">
                                    <div class="category-header" @click="toggleCategory('other')" style="cursor: pointer; display: flex; align-items: center; justify-content: space-between; padding: 0.5rem 0;">
                                        <div style="display: flex; align-items: center; gap: 0.5rem;">
                                            <i :class="['small icon', expandedCategories.other ? 'caret down' : 'caret right']" style="margin: 0; font-size: 0.9em;"></i>
                                            <span class="ui grey text" style="font-weight: 600; font-size: 1.1em;">Other Events</span>
                                        </div>
                                        <span class="ui circular mini label grey">{{ getEventsByCategory('other').length }}</span>
                                    </div>
                                    <div v-show="expandedCategories.other" class="ui three column grid" style="margin-top: 0.75rem; padding-left: 1.5rem;">
                                        <div class="column" v-for="event in getEventsByCategory('other')" :key="event">
                                            <div class="ui checkbox" style="margin-bottom: 0.5rem;">
                                                <input type="checkbox" :id="'event-' + event"
                                                       :value="event" v-model="webhookForm.events">
                                                <label :for="'event-' + event" style="font-size: 0.9em;">{{ formatEventName(event) }}</label>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            </div>

                            <!-- Quick selection buttons -->
                            <div class="ui horizontal divider">Quick Selection</div>
                            <div class="ui buttons mini">
                                <button class="ui button" @click="selectAllEvents">Select All</button>
                                <button class="ui button" @click="deselectAllEvents">Deselect All</button>
                            </div>
                            <div class="ui horizontal list" style="margin-top: 0.5rem;">
                                <div class="item">
                                    <span class="ui green circular label">{{ webhookForm.events.length }}</span>
                                    <span>events selected</span>
                                </div>
                            </div>
                        </div>
                        
                        
                        <div class="field">
                            <label>Description (optional)</label>
                            <textarea v-model="webhookForm.description"
                                      placeholder="Description for this webhook configuration"
                                      rows="2" aria-label="Description"></textarea>
                        </div>
                        
                        <div class="field">
                            <button class="ui primary button" :class="{'loading': loading}"
                                    @click="handleSubmit" type="button" :disabled="loading">
                                {{ editingWebhookId ? 'Update Webhook' : 'Create Webhook' }}
                            </button>
                            <button class="ui button" @click="toggleAddForm" :disabled="loading">
                                Cancel
                            </button>
                        </div>
                    </div>
                </div>
            </transition>
        </div>
        <div class="actions">
            <div class="ui black deny button">
                Close
            </div>
        </div>
    </div>
    `
}