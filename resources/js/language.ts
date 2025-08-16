import { getLocale } from "./lib/helper";

const messages = {
    'zh-CN': {
        // 公共翻译
        common: {
            hello: '你好',
            save: '保存',
            cancel: '取消',
            confirm: '确认',
            delete: '删除',
            edit: '编辑',
            add: '添加',
            search: '搜索',
            loading: '加载中...',
            success: '成功',
            error: '错误',
            warning: '警告',
            info: '信息',
            on: '开启',
            off: '关闭',
            enable: '启用',
            disable: '禁用',
            yes: '是',
            no: '否',
            ok: '确定',
            close: '关闭',
            back: '返回',
            next: '下一步',
            previous: '上一步',
            submit: '提交',
            reset: '重置',
            refresh: '刷新',
            download: '下载',
            upload: '上传',
            copy: '复制',
            paste: '粘贴',
            cut: '剪切',
            select: '选择',
            all: '全部',
            none: '无',
            custom: '自定义',
            default: '默认',
            optional: '可选',
            required: '必填',
            unknown: '未知',
            noData: '暂无数据',
            noResults: '暂无结果',
            networkError: '网络错误',
            serverError: '服务器错误',
            timeout: '请求超时',
            unauthorized: '未授权',
            forbidden: '禁止访问',
            notFound: '页面不存在',
            internalError: '内部错误'
        },

        // 导航
        navigation: {
            home: '首页',
            progress: '进度',
            subscription: '订阅',
            cookie: '会话',
            settings: '设置',
            about: '关于',
            menu: '菜单'
        },

        // 首页
        home: {
            title: '首页',
            welcome: '欢迎使用',
            description: '这是一个功能强大的B站视频管理工具',
            favoriteList: '收藏夹列表',
            created: '创建',
            updated: '更新'
        },

        // 关于页面
        about: {
            title: '关于',
            subtitle: 'bilibili 收藏夹下载工具',
            description: '主要用于解决收藏夹视频消失的问题，帮助用户备份重要的收藏内容。',
            mainFeatures: '主要功能：',
            features: {
                syncFavorites: '定时同步收藏夹视频信息',
                autoDownload: '自动下载高画质视频备份',
                onlinePlayback: '提供在线播放和管理界面',
                danmakuDownload: '自动下载视频弹幕数据'
            },
            viewOnGitHub: '在 GitHub 上查看',
            systemInfo: '系统信息',
            versionInfo: '版本信息',
            timeInfo: '时间信息',
            databaseUsage: '数据库使用情况',
            appVersion: '应用版本',
            phpVersion: 'PHP 版本',
            laravelVersion: 'Laravel 版本',
            databaseVersion: '数据库版本',
            timezone: '时区',
            currentTime: '当前时间',
            favoriteLists: '收藏夹列表',
            videos: '视频数量',
            videoParts: '视频分片',
            danmaku: '弹幕数量',
            databaseSize: '数据库大小',
            mediaVideosUsage: '媒体视频使用情况',
            mediaImagesUsage: '媒体图片使用情况',
            units: {
                count: '个',
                danmaku: '条',
                mb: 'MB'
            }
        },

        // 进度页面
        progress: {
            title: '进度',
            viewTasks: '查看任务',
            cacheRate: '缓存的视频率',
            cacheRateDescription: '如果你的收藏夹中出现了无效视频那么就会低于100%',
            showCachedOnly: '只显示本地缓存的视频',
            allVideos: '所有视频',
            allVideosDescription: '你所有收藏的视频数',
            validVideos: '有效视频',
            validVideosDescription: '目前仍可以在线观看的视频',
            invalidVideos: '无效视频',
            invalidVideosDescription: '收藏的视频无效被下架',
            frozenVideos: '冻结视频',
            frozenVideosDescription: '当你收藏的视频缓存了之后, 如果视频被删除下架那么就会将该视频归纳为冻结',
            published: '发布',
            favorited: '收藏'
        },

        // 收藏夹页面
        favorites: {
            title: '收藏夹',
            sync: '同步收藏夹',
            syncSuccess: '收藏夹同步成功',
            syncError: '收藏夹同步失败',
            empty: '暂无收藏内容',
            loading: '正在加载收藏夹...',
            published: '发布',
            favorited: '收藏',
            downloaded: '已下载'
        },

        // 视频页面
        video: {
            title: '视频',
            videoParts: '视频选集',
            videoDescription: '视频简介',
            publishTime: '发布时间',
            favoriteTime: '收藏时间',
            danmakuCount: '弹幕数量',
            watchOnBilibili: '在哔哩哔哩观看',
            videoNotFound: '视频未找到',
            videoNotFoundDescription: '抱歉，您要查找的视频不存在或已被删除',
            backToHome: '返回首页',
            loading: '加载中...',
            favorite: '收藏夹'
        },

        // 设置页面
        settings: {
            // 页面标题
            title: '设置',
            saveSettings: '保存设置',
            settingsSaved: '设置保存成功！',
            settingsSaveFailed: '设置保存失败！',
            
            // 分组标题
            featureSwitches: '功能开关',
            filterSettings: '过滤设置',
            notificationSettings: '通知设置',
            
            // 功能开关
            features: {
                favoriteSync: '收藏夹同步',
                videoDownload: '视频下载',
                danmakuDownload: '弹幕下载',
                multiPartitionDownload: '多分P下载',
                humanReadableName: '可读文件名',
                usageAnalytics: '使用情况统计',
                usageAnalyticsDescription: '帮助我们了解功能使用情况，改进产品体验（完全匿名）'
            },
            
            // 过滤设置
            filters: {
                byName: '按名称过滤',
                bySize: '按大小过滤',
                byFavorites: '按收藏夹过滤',
                noFilter: '不过滤',
                containsKeyword: '包含关键词',
                regexPattern: '正则表达式',
                largerThan1GB: '大于 1GB',
                largerThan2GB: '大于 2GB',
                customSize: '自定义大小',
                enableFavoritesFilter: '启用收藏夹过滤',
                selectExcludedFavorites: '选择要排除的收藏夹：'
            },
            
            // 输入框占位符
            placeholders: {
                enterKeyword: '输入要过滤的关键词',
                enterRegex: '输入正则表达式',
                enterSizeMB: '输入大小（MB）',
                enterBotToken: '请输入 Bot Token',
                enterChatId: '请输入聊天 ID'
            },
            
            // 通知设置
            notifications: {
                telegramBot: 'Telegram Bot 通知',
                botToken: 'Bot Token',
                chatId: '聊天 ID',
                connectionTest: '连接测试',
                testConnection: '测试连接',
                enableDescription: '启用后可通过 Telegram Bot 接收系统通知',
                botTokenDescription: '从 BotFather 获取的 Bot Token',
                chatIdDescription: '接收通知的聊天 ID（个人或群组）',
                testConnectionDescription: '测试 Telegram Bot 连接是否正常',
                connectionSuccess: '✅ 连接成功！Telegram Bot 配置正确',
                connectionFailed: '❌ 连接失败！请检查 Bot Token 和聊天 ID 是否正确',
                connectionError: '❌ 连接测试失败！请检查网络连接或稍后重试',
                fillRequiredFields: '请先填写 Bot Token 和聊天 ID'
            }
        },

        // 会话页面
        cookie: {
            title: '会话',
            fileStatus: '文件状态',
            exist: '存在',
            notExist: '不存在',
            cookieStatus: 'COOKIE 状态',
            valid: '有效',
            invalid: '无效',
            check: '检查',
            upload: '上传'
        },

        // 下载页面
        download: {
            title: '下载管理',
            queue: '下载队列',
            completed: '已完成',
            failed: '下载失败',
            pause: '暂停',
            resume: '继续',
            cancel: '取消下载',
            retry: '重试',
            clear: '清空队列',
            downloadAll: '下载全部',
            downloadSelected: '下载选中',
            downloadProgress: '下载进度',
            downloadSpeed: '下载速度',
            remainingTime: '剩余时间',
            fileSize: '文件大小',
            downloadPath: '下载路径',
            openFolder: '打开文件夹'
        },

        // 用户相关
        user: {
            profile: '个人资料',
            login: '登录',
            logout: '退出登录',
            register: '注册',
            username: '用户名',
            password: '密码',
            email: '邮箱',
            avatar: '头像',
            settings: '个人设置',
            changePassword: '修改密码',
            forgotPassword: '忘记密码',
            rememberMe: '记住我',
            loginSuccess: '登录成功',
            loginFailed: '登录失败',
            logoutSuccess: '退出成功'
        },

        // 错误页面
        error: {
            notFound: '页面不存在',
            unauthorized: '未授权访问',
            forbidden: '禁止访问',
            serverError: '服务器错误',
            networkError: '网络错误',
            goHome: '返回首页',
            goBack: '返回上页',
            retry: '重试'
        },

        // 订阅管理页面
        subscription: {
            title: '订阅',
            createSubscription: '创建订阅',
            viewSwitch: '视图',
            unifiedView: '统一',
            categorizedView: '分类',
            mixedView: '混合',
            unifiedCardStyle: '统一卡片样式',
            categorizedDisplay: '分类分区展示',
            mixedViewTitle: '混合视图',
            allSubscriptionsOverview: '所有订阅概览',
            seriesSubscriptions: '系列订阅',
            upSubscriptions: 'UP主订阅',
            subscriptionType: '订阅类型',
            subscriptionLink: '订阅链接',
            enterSubscriptionLink: '请输入订阅链接',
            upMaster: 'UP主',
            series: '系列',
            status: '状态',
            enabled: '启用',
            disabled: '关闭',
            videoCount: '个视频',
            upMasterId: 'UP主ID',
            lastCheck: '最后检查',
            lastUpdate: '最后更新',
            description: '描述',
            actions: '操作',
            delete: '删除',
            cannotDelete: '不可删除',
            confirmDelete: '确定要删除这个订阅吗？',
            createNewSubscription: '创建新订阅',
            subscriptionCreated: '订阅创建成功',
            subscriptionDeleted: '订阅删除成功',
            deleteSubscriptionFailed: '删除订阅失败',
            statusToggleSuccess: '状态切换成功',
            statusToggleFailed: '状态切换失败',
            createSubscriptionFailed: '创建订阅失败',
            // 视图相关
            view: {
                unified: '统一视图',
                categorized: '分类视图',
                mixed: '混合视图'
            },
        }
    },

    'en-US': {
        // 公共翻译
        common: {
            hello: 'Hello',
            save: 'Save',
            cancel: 'Cancel',
            confirm: 'Confirm',
            delete: 'Delete',
            edit: 'Edit',
            add: 'Add',
            search: 'Search',
            loading: 'Loading...',
            success: 'Success',
            error: 'Error',
            warning: 'Warning',
            info: 'Info',
            on: 'On',
            off: 'Off',
            enable: 'Enable',
            disable: 'Disable',
            yes: 'Yes',
            no: 'No',
            ok: 'OK',
            close: 'Close',
            back: 'Back',
            next: 'Next',
            previous: 'Previous',
            submit: 'Submit',
            reset: 'Reset',
            refresh: 'Refresh',
            download: 'Download',
            upload: 'Upload',
            copy: 'Copy',
            paste: 'Paste',
            cut: 'Cut',
            select: 'Select',
            all: 'All',
            none: 'None',
            custom: 'Custom',
            default: 'Default',
            optional: 'Optional',
            required: 'Required',
            unknown: 'Unknown',
            noData: 'No Data',
            noResults: 'No Results',
            networkError: 'Network Error',
            serverError: 'Server Error',
            timeout: 'Request Timeout',
            unauthorized: 'Unauthorized',
            forbidden: 'Forbidden',
            notFound: 'Page Not Found',
            internalError: 'Internal Error'
        },

        // 导航
        navigation: {
            home: 'Home',
            progress: 'Progress',
            subscription: 'Subscription',
            cookie: 'Cookie',
            settings: 'Settings',
            about: 'About',
            menu: 'Menu'
        },

        // 首页
        home: {
            title: 'Home',
            welcome: 'Welcome',
            description: 'A powerful Bilibili video management tool',
            favoriteList: 'Favorite List',
            created: 'Created',
            updated: 'Updated'
        },

        // 关于页面
        about: {
            title: 'About',
            subtitle: 'Bilibili Favorite Download Tool',
            description: 'Mainly used to solve the problem of disappearing favorite videos and help users backup important favorite content.',
            mainFeatures: 'Main Features:',
            features: {
                syncFavorites: 'Scheduled synchronization of favorite video information',
                autoDownload: 'Automatic download of high-quality video backups',
                onlinePlayback: 'Provide online playback and management interface',
                danmakuDownload: 'Automatic download of video danmaku data'
            },
            viewOnGitHub: 'View on GitHub',
            systemInfo: 'System Information',
            versionInfo: 'Version Information',
            timeInfo: 'Time Information',
            databaseUsage: 'Database Usage',
            appVersion: 'App Version',
            phpVersion: 'PHP Version',
            laravelVersion: 'Laravel Version',
            databaseVersion: 'Database Version',
            timezone: 'Timezone',
            currentTime: 'Current Time',
            favoriteLists: 'Favorite Lists',
            videos: 'Videos',
            videoParts: 'Video Parts',
            danmaku: 'Danmaku',
            databaseSize: 'Database Size',
            mediaVideosUsage: 'Media Videos Usage',
            mediaImagesUsage: 'Media Images Usage',
            units: {
                count: '',
                danmaku: '',
                mb: 'MB'
            }
        },

        // 进度页面
        progress: {
            title: 'Progress',
            viewTasks: 'View Tasks',
            cacheRate: 'Cached Video Rate',
            cacheRateDescription: 'If invalid videos appear in your favorites, it will be below 100%',
            showCachedOnly: 'Downloaded',
            allVideos: 'All Videos',
            allVideosDescription: 'Total number of videos in your favorites',
            validVideos: 'Valid Videos',
            validVideosDescription: 'Videos that can still be watched online',
            invalidVideos: 'Invalid Videos',
            invalidVideosDescription: 'Favorited videos that have been taken down',
            frozenVideos: 'Frozen Videos',
            frozenVideosDescription: 'When your favorited videos are cached, if the video is deleted or taken down, it will be categorized as frozen',
            published: 'Published',
            favorited: 'Favorited'
        },

        // 收藏夹页面
        favorites: {
            title: 'Favorites',
            sync: 'Sync Favorites',
            syncSuccess: 'Favorites synced successfully',
            syncError: 'Failed to sync favorites',
            empty: 'No favorites yet',
            loading: 'Loading favorites...',
            published: 'Published',
            favorited: 'Favorited',
            downloaded: 'Downloaded'
        },

        // 视频页面
        video: {
            title: 'Video',
            videoParts: 'Video Parts',
            videoDescription: 'Video Description',
            publishTime: 'Publish Time',
            favoriteTime: 'Favorite Time',
            danmakuCount: 'Danmaku Count',
            watchOnBilibili: 'Watch on Bilibili',
            videoNotFound: 'Video Not Found',
            videoNotFoundDescription: 'Sorry, the video you are looking for does not exist or has been deleted',
            backToHome: 'Back to Home',
            loading: 'Loading...',
            favorite: 'Favorite'
        },

        // 设置页面
        settings: {
            // 页面标题
            title: 'Settings',
            saveSettings: 'Save Settings',
            settingsSaved: 'Settings saved successfully!',
            settingsSaveFailed: 'Settings save failed!',
            
            // 分组标题
            featureSwitches: 'Feature Switches',
            filterSettings: 'Filter Settings',
            notificationSettings: 'Notification Settings',
            
            // 功能开关
            features: {
                favoriteSync: 'Favorite Sync',
                videoDownload: 'Video Download',
                danmakuDownload: 'Danmaku Download',
                multiPartitionDownload: 'Multi-Partition Download',
                humanReadableName: 'Human Readable Name',
                usageAnalytics: 'Usage Analytics',
                usageAnalyticsDescription: 'Help us understand feature usage and improve product experience (completely anonymous)'
            },
            
            // 过滤设置
            filters: {
                byName: 'Filter by Name',
                bySize: 'Filter by Size',
                byFavorites: 'Filter by Favorites',
                noFilter: 'No Filter',
                containsKeyword: 'Contains Keyword',
                regexPattern: 'Regex Pattern',
                largerThan1GB: 'Larger than 1GB',
                largerThan2GB: 'Larger than 2GB',
                customSize: 'Custom Size',
                enableFavoritesFilter: 'Enable Favorites Filter',
                selectExcludedFavorites: 'Select favorites to exclude:'
            },
            
            // 输入框占位符
            placeholders: {
                enterKeyword: 'Enter keyword to filter',
                enterRegex: 'Enter regex pattern',
                enterSizeMB: 'Enter size in MB',
                enterBotToken: 'Enter Bot Token',
                enterChatId: 'Enter Chat ID'
            },
            
            // 通知设置
            notifications: {
                telegramBot: 'Telegram Bot Notifications',
                botToken: 'Bot Token',
                chatId: 'Chat ID',
                connectionTest: 'Connection Test',
                testConnection: 'Test Connection',
                enableDescription: 'Enable to receive system notifications via Telegram Bot',
                botTokenDescription: 'Bot Token obtained from BotFather',
                chatIdDescription: 'Chat ID to receive notifications (personal or group)',
                testConnectionDescription: 'Test if Telegram Bot connection is working properly',
                connectionSuccess: '✅ Connection successful! Telegram Bot configuration is correct',
                connectionFailed: '❌ Connection failed! Please check if Bot Token and Chat ID are correct',
                connectionError: '❌ Connection test failed! Please check network connection or try again later',
                fillRequiredFields: 'Please fill in Bot Token and Chat ID first'
            }
        },

        // 会话页面
        cookie: {
            title: 'Cookie',
            fileStatus: 'File Status',
            exist: 'Exist',
            notExist: 'Not Exist',
            cookieStatus: 'Cookie Status',
            valid: 'Valid',
            invalid: 'Invalid',
            check: 'Check',
            upload: 'Upload'
        },

        // 下载页面
        download: {
            title: 'Download Manager',
            queue: 'Download Queue',
            completed: 'Completed',
            failed: 'Failed',
            pause: 'Pause',
            resume: 'Resume',
            cancel: 'Cancel Download',
            retry: 'Retry',
            clear: 'Clear Queue',
            downloadAll: 'Download All',
            downloadSelected: 'Download Selected',
            downloadProgress: 'Download Progress',
            downloadSpeed: 'Download Speed',
            remainingTime: 'Remaining Time',
            fileSize: 'File Size',
            downloadPath: 'Download Path',
            openFolder: 'Open Folder'
        },

        // 用户相关
        user: {
            profile: 'Profile',
            login: 'Login',
            logout: 'Logout',
            register: 'Register',
            username: 'Username',
            password: 'Password',
            email: 'Email',
            avatar: 'Avatar',
            settings: 'Settings',
            changePassword: 'Change Password',
            forgotPassword: 'Forgot Password',
            rememberMe: 'Remember Me',
            loginSuccess: 'Login Successful',
            loginFailed: 'Login Failed',
            logoutSuccess: 'Logout Successful'
        },

        // 错误页面
        error: {
            notFound: 'Page Not Found',
            unauthorized: 'Unauthorized',
            forbidden: 'Forbidden',
            serverError: 'Server Error',
            networkError: 'Network Error',
            goHome: 'Go Home',
            goBack: 'Go Back',
            retry: 'Retry'
        },

        // 订阅管理页面
        subscription: {
            title: 'Subscription',
            createSubscription: 'Create Subscription',
            viewSwitch: 'View',
            unifiedView: 'Unified',
            categorizedView: 'Categorized',
            mixedView: 'Mixed',
            unifiedCardStyle: 'Unified Card Style',
            categorizedDisplay: 'Categorized Partition Display',
            mixedViewTitle: 'Mixed View',
            allSubscriptionsOverview: 'All Subscriptions Overview',
            seriesSubscriptions: 'Series Subscriptions',
            upSubscriptions: 'UP Master Subscriptions',
            subscriptionType: 'Subscription Type',
            subscriptionLink: 'Subscription Link',
            enterSubscriptionLink: 'Please enter the subscription link',
            upMaster: 'UP Master',
            series: 'Series',
            status: 'Status',
            enabled: 'Enabled',
            disabled: 'Disabled',
            videoCount: 'videos',
            upMasterId: 'UP Master ID',
            lastCheck: 'Last Check',
            lastUpdate: 'Last Update',
            description: 'Description',
            actions: 'Actions',
            delete: 'Delete',
            cannotDelete: 'Cannot Delete',
            confirmDelete: 'Are you sure you want to delete this subscription?',
            createNewSubscription: 'Create New Subscription',
            subscriptionCreated: 'Subscription created successfully',
            subscriptionDeleted: 'Subscription deleted successfully',
            deleteSubscriptionFailed: 'Delete subscription failed',
            statusToggleSuccess: 'Status toggled successfully',
            statusToggleFailed: 'Status toggled failed',
            createSubscriptionFailed: 'Create subscription failed',
            // View related
            view: {
                unified: 'Unified View',
                categorized: 'Categorized View',
                mixed: 'Mixed View'
            },
        }
    }
};

export default {
    locale: getLocale(), // 默认语言
    fallbackLocale: 'zh-CN', // 回退语言
    messages
};