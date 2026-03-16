#!/bin/bash

# 交互式批量ETC1转换脚本 (etc1tool在当前目录)

# 检查etc1tool是否存在
if [ ! -f "./etc1tool" ]; then
    echo "错误: etc1tool未在当前目录中找到"
    echo "请将etc1tool放在当前目录: $(pwd)"
    exit 1
fi

echo "ETC1批量转换工具 (etc1tool在当前目录)"
echo "======================================"
echo ""

# 选择模式
echo "请选择操作模式:"
echo "1. 编码PNG到ETC1 (PNG -> ETC1)"
echo "2. 解码ETC1到PNG (ETC1 -> PNG)"
echo "3. 编码PNG到raw ETC1 (无头部)"
echo "4. 批量重命名ETC1文件扩展名"
read -p "请输入选择 (1-4): " choice

case $choice in
    1)
        MODE="encode"
        ENCODE_OPTION="--encode"
        INPUT_EXT="png"
        OUTPUT_EXT="pkm"
        ;;
    2)
        MODE="decode"
        ENCODE_OPTION="--decode"
        INPUT_EXT="pkm"
        OUTPUT_EXT="png"
        ;;
    3)
        MODE="encode"
        ENCODE_OPTION="--encodeNoHeader"
        INPUT_EXT="png"
        OUTPUT_EXT="etc1"
        ;;
    4)
        # 重命名模式
        echo ""
        echo "批量重命名ETC1文件扩展名"
        read -p "输入要重命名的目录: " RENAME_DIR
        read -p "输入原扩展名 (如: etc1): " OLD_EXT
        read -p "输入新扩展名 (如: pkm): " NEW_EXT
        
        if [ ! -d "$RENAME_DIR" ]; then
            echo "错误: 目录不存在"
            exit 1
        fi
        
        count=0
        for file in "$RENAME_DIR"/*.$OLD_EXT; do
            if [ -f "$file" ]; then
                newfile="${file%.$OLD_EXT}.$NEW_EXT"
                mv "$file" "$newfile"
                echo "重命名: $(basename "$file") -> $(basename "$newfile")"
                ((count++))
            fi
        done
        
        echo "完成! 重命名了 $count 个文件"
        exit 0
        ;;
    *)
        echo "无效选择"
        exit 1
        ;;
esac

# 获取输入输出目录
echo ""
read -p "输入源文件目录 (包含.$INPUT_EXT文件): " INPUT_DIR
read -p "输入输出目录: " OUTPUT_DIR

# 检查输入目录
if [ ! -d "$INPUT_DIR" ]; then
    echo "错误: 输入目录不存在"
    exit 1
fi

# 创建输出目录
mkdir -p "$OUTPUT_DIR"

# 是否递归处理
read -p "是否递归处理子目录? (y/n, 默认n): " RECURSIVE
if [[ "$RECURSIVE" == "y" || "$RECURSIVE" == "Y" ]]; then
    FIND_CMD="find \"$INPUT_DIR\" -type f -name \"*.$INPUT_EXT\""
else
    FIND_CMD="find \"$INPUT_DIR\" -maxdepth 1 -type f -name \"*.$INPUT_EXT\""
fi

# 开始处理
echo ""
echo "开始处理..."
count=0

eval "$FIND_CMD" | while read -r input_file; do
    # 获取文件名（不含扩展名）
    filename=$(basename "$input_file" ".$INPUT_EXT")
    
    # 构建输出路径
    if [[ "$RECURSIVE" == "y" || "$RECURSIVE" == "Y" ]]; then
        rel_path="${input_file#$INPUT_DIR/}"
        rel_path="${rel_path%.$INPUT_EXT}"
        output_file="$OUTPUT_DIR/${rel_path}.$OUTPUT_EXT"
        mkdir -p "$(dirname "$output_file")"
    else
        output_file="$OUTPUT_DIR/${filename}.$OUTPUT_EXT"
    fi
    
    echo "处理: $(basename "$input_file") -> $(basename "$output_file")"
    
    # 执行转换
    if ./etc1tool "$input_file" $ENCODE_OPTION -o "$output_file"; then
        echo "  成功"
        ((count++))
    else
        echo "  失败"
    fi
done

echo ""
echo "完成! 处理了 $count 个文件"