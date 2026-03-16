#!/bin/bash

# 定义可执行文件路径
GO_VERSION="./go/etc1tool"
CPP_VERSION="./cpp/etc1tool"

# 测试图像列表
IMAGES=("test1.png" "test2.png" "test3.png")

# 检查可执行文件是否存在
if [ ! -f "$GO_VERSION" ]; then
    echo "Error: Go version executable not found at $GO_VERSION"
    exit 1
fi

if [ ! -f "$CPP_VERSION" ]; then
    echo "Error: C++ version executable not found at $CPP_VERSION"
    exit 1
fi

# 检查测试图像是否存在
for img in "${IMAGES[@]}"; do
    if [ ! -f "$img" ]; then
        echo "Warning: Test image $img not found"
    fi
done

# 测试所有图像
for img in "${IMAGES[@]}"; do
    if [ ! -f "$img" ]; then
        continue
    fi
    
    echo "Testing $img..."
    
    # 编码测试 - 标准ETC1
    "$GO_VERSION" "$img" --encode -o "${img%.png}_go.pkm"
    "$CPP_VERSION" "$img" --encode -o "${img%.png}_cpp.pkm"
    
    # 比较输出
    if cmp "${img%.png}_go.pkm" "${img%.png}_cpp.pkm"; then
        echo "✓ Standard ETC1 encoding match for $img"
    else
        echo "✗ Standard ETC1 encoding mismatch for $img"
    fi
    
    # 编码测试 - ETC1S
    "$GO_VERSION" "$img" --encodeETC1S -o "${img%.png}_go_etc1s.pkm"
    "$CPP_VERSION" "$img" --encodeETC1S -o "${img%.png}_cpp_etc1s.pkm"
    
    # 比较输出
    if cmp "${img%.png}_go_etc1s.pkm" "${img%.png}_cpp_etc1s.pkm"; then
        echo "✓ ETC1S encoding match for $img"
    else
        echo "✗ ETC1S encoding mismatch for $img"
    fi
    
    # 解码测试 - Go版本解码C++编码的文件
    "$GO_VERSION" "${img%.png}_cpp.pkm" --decode -o "${img%.png}_go_decode.png"
    
    # 解码测试 - C++版本解码Go编码的文件
    "$CPP_VERSION" "${img%.png}_go.pkm" --decode -o "${img%.png}_cpp_decode.png"
    
    # 比较解码输出
    if cmp "${img%.png}_go_decode.png" "${img%.png}_cpp_decode.png"; then
        echo "✓ Decoding match for $img"
    else
        echo "✗ Decoding mismatch for $img"
    fi
    
    echo
done

# 清理临时文件
echo "Cleaning up temporary files..."
for img in "${IMAGES[@]}"; do
    if [ ! -f "$img" ]; then
        continue
    fi
    
    rm -f "${img%.png}_go.pkm"
    rm -f "${img%.png}_cpp.pkm"
    rm -f "${img%.png}_go_etc1s.pkm"
    rm -f "${img%.png}_cpp_etc1s.pkm"
    rm -f "${img%.png}_go_decode.png"
    rm -f "${img%.png}_cpp_decode.png"
done

echo "Test completed!"
