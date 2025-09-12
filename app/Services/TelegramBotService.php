<?php

namespace App\Services;

use App\Contracts\TelegramBotServiceInterface;
use App\Enums\SettingKey;
use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Log;

class TelegramBotService implements TelegramBotServiceInterface
{

    public function __construct(
        private SettingsService $settings
    ) {
    }

    /**
     * 发送消息到Telegram
     * 
     * @param string $message 要发送的消息
     * @param bool $silent 是否静默发送（不显示通知声音）
     * @return bool 发送是否成功
     */
    public function sendMessage(string $message, bool $silent = false): bool
    {
        // 检查是否启用
        if (!$this->isEnabled()) {
            return false;
        }

        $botToken = $this->getBotToken();
        $chatId = $this->getChatId();
        $botUrl = $this->getBotUrl();
        if (!$botToken || !$chatId) {
            Log::warning('Telegram bot not configured properly', [
                'bot_token' => $botToken ? 'set' : 'not_set',
                'chat_id' => $chatId ? 'set' : 'not_set'
            ]);
            return false;
        }

        try {
            $url = "$botUrl/bot{$botToken}/sendMessage";
            
            $data = [
                'chat_id' => $chatId,
                'text' => $message,
                'parse_mode' => 'HTML', // 支持HTML格式
            ];

            if ($silent) {
                $data['disable_notification'] = true;
            }

            $response = Http::timeout(10)->post($url, $data);

            if ($response->successful()) {
                $result = $response->json();
                if ($result['ok'] ?? false) {
                    Log::info('Telegram message sent successfully', [
                        'chat_id' => $chatId,
                        'message_length' => strlen($message)
                    ]);
                    return true;
                }
            }

            Log::error('Failed to send Telegram message', [
                'status_code' => $response->status(),
                'response' => $response->body(),
                'chat_id' => $chatId
            ]);

            return false;

        } catch (\Exception $e) {
            Log::error('Exception while sending Telegram message', [
                'error' => $e->getMessage(),
                'chat_id' => $chatId
            ]);
            return false;
        }
    }

    /**
     * 发送HTML格式的消息
     * 
     * @param string $htmlMessage HTML格式的消息
     * @param bool $silent 是否静默发送
     * @return bool 发送是否成功
     */
    public function sendHtmlMessage(string $htmlMessage, bool $silent = false): bool
    {
        return $this->sendMessage($htmlMessage, $silent);
    }

    /**
     * 发送Markdown格式的消息
     * 
     * @param string $markdownMessage Markdown格式的消息
     * @param bool $silent 是否静默发送
     * @return bool 发送是否成功
     */
    public function sendMarkdownMessage(string $markdownMessage, bool $silent = false): bool
    {
        if (!$this->isEnabled()) {
            return false;
        }

        $botToken = $this->getBotToken();
        $chatId = $this->getChatId();
        $botUrl = $this->getBotUrl();

        if (!$botToken || !$chatId) {
            return false;
        }

        try {
            $url = "$botUrl/bot{$botToken}/sendMessage";
            
            $data = [
                'chat_id' => $chatId,
                'text' => $markdownMessage,
                'parse_mode' => 'MarkdownV2',
            ];

            if ($silent) {
                $data['disable_notification'] = true;
            }

            $response = Http::timeout(10)->post($url, $data);

            if ($response->successful()) {
                $result = $response->json();
                if ($result['ok'] ?? false) {
                    Log::info('Telegram markdown message sent successfully', [
                        'chat_id' => $chatId,
                        'message_length' => strlen($markdownMessage)
                    ]);
                    return true;
                }
            }

            Log::error('Failed to send Telegram markdown message', [
                'status_code' => $response->status(),
                'response' => $response->body(),
                'chat_id' => $chatId
            ]);

            return false;

        } catch (\Exception $e) {
            Log::error('Exception while sending Telegram markdown message', [
                'error' => $e->getMessage(),
                'chat_id' => $chatId
            ]);
            return false;
        }
    }

    /**
     * 检查Telegram Bot是否启用
     * 
     * @return bool
     */
    public function isEnabled(): bool
    {
        return $this->settings->get(SettingKey::TELEGRAM_BOT_ENABLED->value) == 'on';
    }

    /**
     * 获取Bot Token
     * 
     * @return string|null
     */
    public function getBotToken(): ?string
    {
        return $this->settings->get(SettingKey::TELEGRAM_BOT_TOKEN->value);
    }

    /**
     * 获取Chat ID
     * 
     * @return string|null
     */
    public function getChatId(): ?string
    {
        return $this->settings->get(SettingKey::TELEGRAM_CHAT_ID->value);
    }

    /**
     * 获取Bot URL
     * 
     * @return string|null
     */
    public function getBotUrl(): ?string
    {
        $url = $this->settings->get(SettingKey::TELEGRAM_BOT_API_URL->value);
        if(empty($url)){
            return 'https://api.telegram.org';
        }
        return $url;
    }

    /**
     * 设置Bot Token
     * 
     * @param string $token
     * @return void
     */
    public function setBotToken(string $token): void
    {
        $this->settings->put(SettingKey::TELEGRAM_BOT_TOKEN->value, $token);
    }

    /**
     * 设置Chat ID
     * 
     * @param string $chatId
     * @return void
     */
    public function setChatId(string $chatId): void
    {
        $this->settings->put(SettingKey::TELEGRAM_CHAT_ID->value, $chatId);
    }

    /**
     * 启用或禁用Telegram Bot
     * 
     * @param bool $enabled
     * @return void
     */
    public function setEnabled(bool $enabled): void
    {
        $this->settings->put(SettingKey::TELEGRAM_BOT_ENABLED->value, $enabled);
    }

    /**
     * 设置Bot URL
     * 
     * @param string $url
     * @return void
     */
    public function setBotUrl(string $url): void
    {
        $this->settings->put(SettingKey::TELEGRAM_BOT_API_URL->value, $url);
    }

    /**
     * 测试Bot连接
     * 
     * @return bool
     */
    public function testConnection(array $args = []): bool
    {
        if(!empty($args)){
            $botToken = $args['bot_token'];
            $chatId = $args['chat_id'];
            $botUrl = $args['bot_url'] ?? 'https://api.telegram.org';
        }else{
            $botToken = $this->getBotToken();
            $chatId = $this->getChatId();
            $botUrl = $this->getBotUrl();
        }

        if (!$botToken || !$chatId) {
            return false;
        }

        try {
            $url = "$botUrl/bot{$botToken}/getMe";
            Log::info('Testing Telegram bot connection', [
                'url' => $url,
                'args' => $args,
            ]);
            $response = Http::timeout(10)->get($url);

            if ($response->successful()) {
                $result = $response->json();
                if ($result['ok'] ?? false) {
                    // 发送测试消息
                    $testMessage = "Test Message - Your Telegram Bot has been connected successfully!";
                    $sendUrl = "$botUrl/bot{$botToken}/sendMessage";
                    $sendResponse = Http::post($sendUrl, [
                        'chat_id' => $chatId,
                        'text' => $testMessage
                    ]);
                    
                    return $sendResponse->successful();
                }
            }

            return false;

        } catch (\Exception $e) {
            Log::error('Exception while testing Telegram bot connection', [
                'error' => $e->getMessage()
            ]);
            return false;
        }
    }
} 