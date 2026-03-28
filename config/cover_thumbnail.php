<?php

return [

    'max_width' => (int) env('COVER_THUMB_MAX_WIDTH', 320),

    'max_height' => (int) env('COVER_THUMB_MAX_HEIGHT', 320),

    /** 定时任务每批最多处理条数 */
    'batch_limit' => (int) env('COVER_THUMB_BATCH_LIMIT', 500),

];
