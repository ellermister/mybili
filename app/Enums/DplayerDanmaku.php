<?php

namespace App\Enums;

class DplayerDanmaku
{
    const RIGHT = 0;
    const TOP = 1;
    const BOTTOM = 2;

    public static function getMode($mode)
    {
        switch ($mode) {
            case 1:
                return self::RIGHT;
            case 4:
                return self::BOTTOM;
            case 5:
                return self::TOP;
            default:
                return self::RIGHT;
        }
    }

}