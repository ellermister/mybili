<?php

namespace App\Contracts;

interface TelegramBotServiceInterface
{
    /**
     * 发送消息到Telegram
     * 
     * @param string $message 要发送的消息
     * @param bool $silent 是否静默发送（不显示通知声音）
     * @return bool 发送是否成功
     */
    public function sendMessage(string $message, bool $silent = false): bool;

    /**
     * 发送HTML格式的消息
     * 
     * @param string $htmlMessage HTML格式的消息
     * @param bool $silent 是否静默发送
     * @return bool 发送是否成功
     */
    public function sendHtmlMessage(string $htmlMessage, bool $silent = false): bool;

    /**
     * 发送Markdown格式的消息
     * 
     * @param string $markdownMessage Markdown格式的消息
     * @param bool $silent 是否静默发送
     * @return bool 发送是否成功
     */
    public function sendMarkdownMessage(string $markdownMessage, bool $silent = false): bool;

    /**
     * 检查Telegram Bot是否启用
     * 
     * @return bool
     */
    public function isEnabled(): bool;

    /**
     * 测试Bot连接
     * 
     * @param array $args 参数
     * @return bool
     */
    public function testConnection(array $args = []): bool;
} 