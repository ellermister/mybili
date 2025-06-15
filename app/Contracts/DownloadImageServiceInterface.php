<?php
namespace App\Contracts;

interface DownloadImageServiceInterface
{
    public function downloadImage(string $url, string $savePath): void;
    public function convertToFilename(string $url): string;
    public function getImagesDirIfNotExistCreate(): string;
    public function getImageLocalPath(string $url): string;
}
